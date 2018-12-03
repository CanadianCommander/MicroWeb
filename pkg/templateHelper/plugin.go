package templateHelper

import (
	"errors"
	templateHTML "html/template"
	"plugin"

	"github.com/CanadianCommander/MicroWeb/pkg/cache"
	"github.com/CanadianCommander/MicroWeb/pkg/logger"
	mwsettings "github.com/CanadianCommander/MicroWeb/pkg/mwSettings"
)

// cache object type string
const templateHelperCacheItem = "templateHelperPlugin:"

type iTemplatePlugin interface {
	getTemplate(t *templateHTML.Template) interface{}
}

type templatePlugin struct {
	getTemplateFunc func(t *templateHTML.Template) interface{}
}

func (tp *templatePlugin) getTemplate(t *templateHTML.Template) interface{} {
	return tp.getTemplateFunc(t)
}

func getPlugin(name string) (iTemplatePlugin, error) {
	plugin := cache.FetchFromCache(templateHelperCacheItem, name)

	if plugin != nil {
		logger.LogVerbose("Loading template helper plugin, %s, from cache", name)
		return plugin.(iTemplatePlugin), nil
	}

	logger.LogVerbose("Loading template helper plugin, %s, from file", name)

	pluginPath, err := pathFromTemplateName(name)
	if err != nil {
		return nil, err
	}
	newPlugin, err := loadPluginFromFile(pluginPath)
	if err != nil {
		return nil, err
	}

	cache.AddToCacheTTLOverride(templateHelperCacheItem, name, cache.MaxTTL, newPlugin)
	return newPlugin, nil

}

func pathFromTemplateName(name string) (string, error) {
	tList := mwsettings.GetSetting("templateHelper/plugins").([]TemplatePluginSettings)
	for _, it := range tList {
		if it.TemplateName == name {
			return it.PluginPath, nil
		}
	}
	return "", errors.New("No template plugin for name: " + name)
}

func loadPluginFromFile(path string) (iTemplatePlugin, error) {
	newTemplatePlugin := templatePlugin{}
	rawPlugin, pErr := plugin.Open(path)

	if pErr != nil {
		return nil, pErr
	}

	newFunc, pErr := rawPlugin.Lookup("GetTemplate")
	if pErr != nil {
		return nil, pErr
	}

	var bOk bool
	newTemplatePlugin.getTemplateFunc, bOk = newFunc.(func(t *templateHTML.Template) interface{})
	if !bOk {
		return nil, errors.New("GetTemplate() is not of correct format. should be: func(t *templateHTML.Template) interface{}")
	}

	return &newTemplatePlugin, nil
}
