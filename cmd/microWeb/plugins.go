package main

import (
	"errors"
	"net/http"
	"path"
	"plugin"
	"sort"
	"time"

	"github.com/CanadianCommander/MicroWeb/pkg/cache"
	"github.com/CanadianCommander/MicroWeb/pkg/logger"
)

type IPlugin interface {
	//called to handle normal resource requests
	HandleRequest(req *http.Request, res http.ResponseWriter, fileContent *[]byte) bool

	//called to handle virtual resource requests (a request the does not target a physical file on the server)
	HandleVirtualRequest(req *http.Request, res http.ResponseWriter) bool
}

type BasicPlugin struct {
	HandleRequestFunc        func(req *http.Request, res http.ResponseWriter, fileContent *[]byte) bool
	HandleVirtualRequestFunc func(req *http.Request, res http.ResponseWriter) bool
}

func (tp *BasicPlugin) HandleRequest(req *http.Request, res http.ResponseWriter, fileContent *[]byte) bool {
	return tp.HandleRequestFunc(req, res, fileContent)
}
func (tp *BasicPlugin) HandleVirtualRequest(req *http.Request, res http.ResponseWriter) bool {
	return tp.HandleVirtualRequestFunc(req, res)
}

func DefaultHandleVirtualRequest(req *http.Request, res http.ResponseWriter) bool {
	res.WriteHeader(404)
	return false
}

func LoadPlugin(path string) (IPlugin, error) {
	cachePlugin := cache.FetchFromCache(cache.CACHE_PLUGIN, path)
	if cachePlugin != nil {
		//cache hit
		plugin := cachePlugin.(IPlugin)
		logger.LogVerbose("Loading plugin from cache: %s", path)
		return plugin, nil
	} else {
		//cache miss
		logger.LogVerbose("Loading plugin from file: %s", path)
		rawPlugin, pError := _loadPlugin(path)
		if pError != nil {
			return nil, pError
		}

		plugin := constructPlugin(rawPlugin)
		if plugin == nil {
			return nil, errors.New("plugin has incorrect format")
		}

		//NOTE, time.Duration(^(uint64(1) << 63)) sets the ttl of plugins to 290 years ... aka never delete
		//really should be a MAX_DURATION type constant. If it exists I couldn't find it.
		cache.AddToCacheTTLOverride(cache.CACHE_PLUGIN, path, time.Duration(^(uint64(1) << 63)), plugin)
		return plugin, nil
	}
}

func _loadPlugin(path string) (*plugin.Plugin, error) {
	newPlugin, pluginError := plugin.Open(path)

	if pluginError != nil {
		logger.LogError("Could not load plugin: %s with Error: %s", path, pluginError.Error())
		return nil, pluginError
	}

	return newPlugin, nil
}

func constructPlugin(plugin *plugin.Plugin) IPlugin {
	NewPlugin := BasicPlugin{}

	handleReqFunc, err := plugin.Lookup("HandleRequest")
	if err != nil {
		logger.LogError("Plugin does not export required function 'func HandleRequest(req *http.Request, res http.ResponseWriter, fileContent *[]byte) bool'")
		return nil
	}
	handleVirtualReqFunc, err := plugin.Lookup("HandleVirtualRequest")
	if err != nil {
		logger.LogInfo("Plugin does not export optional function 'func HandleVirtualRequest(req *http.Request, res http.ResponseWriter) bool, using default'")
		handleVirtualReqFunc = DefaultHandleVirtualRequest
	}

	var bErr bool
	NewPlugin.HandleRequestFunc, bErr = handleReqFunc.(func(req *http.Request, res http.ResponseWriter, fileContent *[]byte) bool)
	if !bErr {
		logger.LogError("Plugin HandleRequest(...) function does not match IPlugin interface")
	}
	NewPlugin.HandleVirtualRequestFunc, bErr = handleVirtualReqFunc.(func(req *http.Request, res http.ResponseWriter) bool)
	if !bErr {
		logger.LogError("Plugin HandleVirtualRequest(...) function does not match IPlugin interface")
	}

	return &NewPlugin
}

//returns the path of the plugin that should be used on the given resource path
func GetPluginByResourcePath(fsPath string) (string, error) {
	pluginList := make([]pluginBinding, len(globalSettings.GetPluginList()))
	copy(pluginList[:], globalSettings.GetPluginList()[:])

	lessFunction := func(i, j int) bool {
		iDist := StringMatchLength(path.Join(globalSettings.GetStaticResourcePath(), pluginList[i].Binding), fsPath)
		jDist := StringMatchLength(path.Join(globalSettings.GetStaticResourcePath(), pluginList[j].Binding), fsPath)
		return iDist > jDist
	}
	sort.Slice(pluginList[:], lessFunction)

	for _, plugin := range pluginList {
		if StringMatchLength(path.Join(globalSettings.GetStaticResourcePath(), plugin.Binding), fsPath) == len(path.Join(globalSettings.GetStaticResourcePath(), plugin.Binding)) {
			return plugin.Plugin, nil
		}
	}

	return "FOOBAR", errors.New("No Plugin found for given path: " + fsPath)
}
