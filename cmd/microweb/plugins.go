package main

import (
	"errors"
	"mime"
	"net/http"
	"path"
	"plugin"
	"sort"

	"github.com/CanadianCommander/MicroWeb/pkg/cache"
	"github.com/CanadianCommander/MicroWeb/pkg/logger"
)

/*
IPlugin a generic plugin interface that all plugins must implement (some functions optional)
*/
type IPlugin interface {
	//called to handle normal resource requests
	HandleRequest(req *http.Request, res http.ResponseWriter, fsName string) bool

	//called to handle virtual resource requests (a request the does not target a physical file on the server)
	HandleVirtualRequest(req *http.Request, res http.ResponseWriter) bool
}

/*
BasicPlugin a basic implementation of the IPlugin interface.
*/
type BasicPlugin struct {
	HandleRequestFunc        func(req *http.Request, res http.ResponseWriter, fsName string) bool
	HandleVirtualRequestFunc func(req *http.Request, res http.ResponseWriter) bool
}

/*
HandleRequest passes through the function call to a function pointer loaded from the plugins symbol table
*/
func (tp *BasicPlugin) HandleRequest(req *http.Request, res http.ResponseWriter, fsName string) bool {
	return tp.HandleRequestFunc(req, res, fsName)
}

/*
HandleVirtualRequest passes through the function call to a function pointer loaded from the plugins symbol table
*/
func (tp *BasicPlugin) HandleVirtualRequest(req *http.Request, res http.ResponseWriter) bool {
	return tp.HandleVirtualRequestFunc(req, res)
}

func defaultHandleVirtualRequest(req *http.Request, res http.ResponseWriter) bool {
	res.WriteHeader(404)
	return false
}

func defaultHandleRequest(req *http.Request, res http.ResponseWriter, fsName string) bool {
	//just serve the file up as is.
	buff := ReadFileToBuff(fsName)
	if buff != nil {
		mimeType := mime.TypeByExtension(path.Ext(fsName))
		res.Header().Add("Content-Type", mimeType)
		res.Write((*buff)[:])
		return true
	}

	res.WriteHeader(500)
	return false
}

/*
LoadPlugin loades the plugin found at path and constructs it.
The constructed plugin is returned as an IPlugin interface on success.

Note. This function uses caching. Therfore you may receive a pointer to an
already initialized plugin
*/
func LoadPlugin(path string) (IPlugin, error) {
	cachePlugin := cache.FetchFromCache(cache.CacheTypePlugin, path)
	if cachePlugin != nil {
		//cache hit
		plugin := cachePlugin.(IPlugin)
		logger.LogVerbose("Loading plugin from cache: %s", path)
		return plugin, nil
	}

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
	cache.AddToCacheTTLOverride(cache.CacheTypePlugin, path, cache.MaxTTL, plugin)
	return plugin, nil

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
		logger.LogInfo("Plugin does not export Optional function 'func HandleRequest(req *http.Request, res http.ResponseWriter, fsName string) bool'")
		handleReqFunc = defaultHandleRequest
	}
	handleVirtualReqFunc, err := plugin.Lookup("HandleVirtualRequest")
	if err != nil {
		logger.LogInfo("Plugin does not export optional function 'func HandleVirtualRequest(req *http.Request, res http.ResponseWriter) bool, using default'")
		handleVirtualReqFunc = defaultHandleVirtualRequest
	}

	var bErr bool
	NewPlugin.HandleRequestFunc, bErr = handleReqFunc.(func(req *http.Request, res http.ResponseWriter, fsName string) bool)
	if !bErr {
		logger.LogError("Plugin HandleRequest(...) function does not match IPlugin interface")
	}
	NewPlugin.HandleVirtualRequestFunc, bErr = handleVirtualReqFunc.(func(req *http.Request, res http.ResponseWriter) bool)
	if !bErr {
		logger.LogError("Plugin HandleVirtualRequest(...) function does not match IPlugin interface")
	}

	return &NewPlugin
}

/*
GetPluginByResourcePath returns the path of the plugin that has the longest binding
match with the given fsPath or an error if no plugin matches at all.

Longest match means, given these two bindings
/index/
/index/web.html
the second binding will be selected for fsPath=/index/web.html because it is longer than
the match produced by the /index/ binding, while for all othere querys, ex fsPath=/index/foo.html
the frist binding will be used.
*/
func GetPluginByResourcePath(fsPath string) (string, error) {
	pluginList := make([]pluginBinding, len(GlobalSettings.GetPluginList()))
	copy(pluginList[:], GlobalSettings.GetPluginList()[:])

	lessFunction := func(i, j int) bool {
		iDist := StringMatchLength(path.Join(GlobalSettings.GetStaticResourcePath(), pluginList[i].Binding), fsPath)
		jDist := StringMatchLength(path.Join(GlobalSettings.GetStaticResourcePath(), pluginList[j].Binding), fsPath)
		return iDist > jDist
	}
	sort.Slice(pluginList[:], lessFunction)

	for _, plugin := range pluginList {
		if StringMatchLength(path.Join(GlobalSettings.GetStaticResourcePath(), plugin.Binding), fsPath) == len(path.Join(GlobalSettings.GetStaticResourcePath(), plugin.Binding)) {
			return plugin.Plugin, nil
		}
	}

	return "", errors.New("No Plugin found for given path: " + fsPath)
}
