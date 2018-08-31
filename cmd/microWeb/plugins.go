package main

import (
	"errors"
	"net/http"
	"path"
	"plugin"
	"sort"

	"github.com/CanadianCommander/MicroWeb/pkg/logger"
)

const (
	TEMPLATE_PLUGIN = iota
	API_PLUGIN
)

type ITemplatePlugin interface {
	//called once to initialize the plugin.
	Init()
	//called before template parsing. this function should return a template structure to be used in said parsing.
	// if this function returns error the entire request is aborted giving the user a 500 (internal server error)
	GetTemplateStruct(r *http.Request) (interface{}, error)
}

type TemplatePlugin struct {
	InitFunc              func()
	GetTemplateStructFunc func(r *http.Request) (interface{}, error)
}

func (tp *TemplatePlugin) Init() {
	tp.InitFunc()
}
func (tp *TemplatePlugin) GetTemplateStruct(r *http.Request) (interface{}, error) {
	return tp.GetTemplateStructFunc(r)
}

func DefaultTemplatePluginInit() {}

func LoadTemplatePlugin(path string) (ITemplatePlugin, error) {
	cachePlugin := FetchFromCache(CACHE_TEMPLATE_PLUGIN, path)
	if cachePlugin != nil {
		//cache hit
		templatePlugin := cachePlugin.(ITemplatePlugin)
		logger.LogVerbose("Loading template plugin from cache: %s", path)
		return templatePlugin, nil
	} else {
		//cache miss
		logger.LogVerbose("Loading template plugin from file: %s", path)
		plugin, pError := LoadPlugin(path)
		if pError != nil {
			return nil, pError
		}

		templatePlugin := ConstructTemplatePlugin(plugin)
		if templatePlugin == nil {
			return nil, errors.New("plugin has incorrect format")
		}

		AddToCache(CACHE_TEMPLATE_PLUGIN, path, templatePlugin)
		return templatePlugin, nil
	}
}

func LoadPlugin(path string) (*plugin.Plugin, error) {
	newPlugin, pluginError := plugin.Open(path)

	if pluginError != nil {
		logger.LogError("Could not load plugin: %s with Error: %s", path, pluginError.Error())
		return nil, pluginError
	}

	return newPlugin, nil
}

func ConstructTemplatePlugin(plugin *plugin.Plugin) ITemplatePlugin {
	tPlug := TemplatePlugin{}

	initFunc, err := plugin.Lookup("Init")
	if err != nil {
		logger.LogWarning("Template plugin does not export optional function Init(), using default")
		initFunc = DefaultTemplatePluginInit
	}
	getTemplateFunc, err := plugin.Lookup("GetTemplateStruct")
	if err != nil {
		logger.LogError("Template plugin does not export required function 'func GetTemplateStruct(r *http.Request) (interface{}, error)'")
		return nil
	}

	var bErr bool
	tPlug.InitFunc, bErr = initFunc.(func())
	if !bErr {
		logger.LogError("Template plugin Init() function does not match template ITemplatePlugin")
		return nil
	}
	tPlug.GetTemplateStructFunc, bErr = getTemplateFunc.(func(r *http.Request) (interface{}, error))
	if !bErr {
		logger.LogError("Template plugin GetTemplateStruct() function does not match ITemplatePlugin")
	}

	return &tPlug
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
