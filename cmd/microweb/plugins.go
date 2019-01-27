package main

import (
	"errors"
	"mime"
	"net/http"
	"path"
	"plugin"
	"reflect"
	"sort"
	"time"

	"github.com/CanadianCommander/MicroWeb/pkg/cache"
	"github.com/CanadianCommander/MicroWeb/pkg/logger"
	mwsettings "github.com/CanadianCommander/MicroWeb/pkg/mwSettings"
)

/*
IPlugin a generic plugin interface that all plugins must implement (some functions optional)
*/
type IPlugin interface {
	//called to intialize the plugin.
	Init()

	//called to handle normal resource requests
	HandleRequest(req *http.Request, res http.ResponseWriter, fsName string) bool

	//called to handle virtual resource requests (a request the does not target a physical file on the server)
	HandleVirtualRequest(req *http.Request, res http.ResponseWriter) bool
}

/*
BasicPlugin a basic implementation of the IPlugin interface.
*/
type BasicPlugin struct {
	bIsInti                  bool
	InitFunc                 func()
	HandleRequestFunc        func(req *http.Request, res http.ResponseWriter, fsName string) bool
	HandleVirtualRequestFunc func(req *http.Request, res http.ResponseWriter) bool
}

/*
Init calls the init function provided by the plugins symbol table.
*/
func (tp *BasicPlugin) Init() {
	tp.InitFunc()
	tp.bIsInti = true
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

func defaultInit() {
	//nop
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
LoadAllPlugins loads all the plugins in the configuration file and calls their init methods.
*/
func LoadAllPlugins() {
	if mwsettings.HasSetting("plugin/plugins") {
		startTime := time.Now()
		logger.LogInfo("loading plugins....")
		pluginList := mwsettings.GetSetting("plugin/plugins").([]pluginBinding)

		for _, plugin := range pluginList {
			_, err := LoadPlugin(plugin.Plugin)
			if err != nil {
				logger.LogError("failed to load plugin with error: %s", err)
			}
			logger.LogVerbose("plugin: %s loaded", plugin.Plugin)
		}
		logger.LogInfo("plugins loaded in %d ms", time.Since(startTime)/time.Millisecond)
	}
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

	//initialize
	plugin.Init()

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

	initFunc, err := plugin.Lookup("Init")
	if err != nil {
		logger.LogInfo("Plugin does not export Optional function 'func Init()', using default")
		initFunc = defaultInit
	}
	handleReqFunc, err := plugin.Lookup("HandleRequest")
	if err != nil {
		logger.LogInfo("Plugin does not export Optional function 'func HandleRequest(req *http.Request, res http.ResponseWriter, fsName string) bool', using default")
		handleReqFunc = defaultHandleRequest
	}
	handleVirtualReqFunc, err := plugin.Lookup("HandleVirtualRequest")
	if err != nil {
		logger.LogInfo("Plugin does not export optional function 'func HandleVirtualRequest(req *http.Request, res http.ResponseWriter) bool', using default")
		handleVirtualReqFunc = defaultHandleVirtualRequest
	}

	var bOk bool
	NewPlugin.InitFunc, bOk = initFunc.(func())
	if !bOk {
		logger.LogError("Plugin Init() function does not match IPlugin interface")
		return nil
	}
	NewPlugin.HandleRequestFunc, bOk = handleReqFunc.(func(req *http.Request, res http.ResponseWriter, fsName string) bool)
	if !bOk {
		logger.LogError("Plugin HandleRequest(...) function does not match IPlugin interface")
		return nil
	}
	NewPlugin.HandleVirtualRequestFunc, bOk = handleVirtualReqFunc.(func(req *http.Request, res http.ResponseWriter) bool)
	if !bOk {
		logger.LogError("Plugin HandleVirtualRequest(...) function does not match IPlugin interface")
		return nil
	}

	return &NewPlugin
}

//pluginBinding represents the plugin setting structure in the configuration function
type pluginBinding struct {
	Binding string
	Plugin  string
}

//AddPluginSettingDecoder adds a decoder for the plugin setting format in the config file.
func AddPluginSettingDecoder() {
	var pluginPath = "plugin/plugins"

	mwsettings.AddSettingDecoder(mwsettings.NewFunctionalSettingDecoder(func(s interface{}) (string, interface{}) {
		if reflect.ValueOf(s).Type().Kind() == reflect.Slice {
			pluginList := s.([]interface{})
			outList := make([]pluginBinding, len(pluginList))

			for i, plugin := range pluginList {
				outList[i] = pluginBinding{}
				outList[i].Binding = plugin.(map[string]interface{})["binding"].(string)
				outList[i].Plugin = plugin.(map[string]interface{})["plugin"].(string)
			}
			return pluginPath, outList
		}

		logger.LogError("Error parsing plugin list. format incorrect")
		return "ERROR", nil
	},
		func(path string) bool {
			if path == pluginPath {
				return true
			}
			return false
		}))
}

/*
GetPluginByResourcePath returns the path of the plugin that has the longest binding
match with the given fsPath or an error if no plugin matches at all.

Longest match means, given these two bindings
/index/
/index/web.html
the second binding will be selected for fsPath=/index/web.html because it is longer than
the match produced by the /index/ binding, while for all other queries, ex fsPath=/index/foo.html
the frist binding will be used.
*/
func GetPluginByResourcePath(fsPath string) (string, error) {
	if mwsettings.HasSetting("plugin/plugins") {
		pluginList := mwsettings.GetSetting("plugin/plugins").([]pluginBinding)

		lessFunction := func(i, j int) bool {
			iDist := StringMatchLength(path.Join(mwsettings.GetSettingString("general/staticDirectory"), pluginList[i].Binding), fsPath)
			jDist := StringMatchLength(path.Join(mwsettings.GetSettingString("general/staticDirectory"), pluginList[j].Binding), fsPath)
			return iDist > jDist
		}
		sort.Slice(pluginList[:], lessFunction)

		for _, plugin := range pluginList {
			if StringMatchLength(path.Join(mwsettings.GetSettingString("general/staticDirectory"), plugin.Binding), fsPath) ==
				len(path.Join(mwsettings.GetSettingString("general/staticDirectory"), plugin.Binding)) {
				return plugin.Plugin, nil
			}
		}
	}

	return "", errors.New("No Plugin found for given path: " + fsPath)
}
