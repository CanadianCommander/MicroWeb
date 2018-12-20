package templateHelper

import (
	"errors"
	templateHTML "html/template"
	"plugin"

	"github.com/CanadianCommander/MicroWeb/pkg/cache"
	"github.com/CanadianCommander/MicroWeb/pkg/logger"
	mwsettings "github.com/CanadianCommander/MicroWeb/pkg/mwSettings"
)

type iTemplatePlugin interface {
	getTemplate(t *templateHTML.Template)
	getData(argv interface{}) interface{}
}

type templatePlugin struct {
	getTemplateFunc func(t *templateHTML.Template)
	getDataFunc     func(argv interface{}) interface{}
}

func (tp *templatePlugin) getTemplate(t *templateHTML.Template) {
	tp.getTemplateFunc(t)
}

func (tp *templatePlugin) getData(argv interface{}) interface{} {
	return tp.getDataFunc(argv)
}

func getPlugin(name string) (iTemplatePlugin, error) {
	plugin := cache.FetchFromCache(cache.CacheTypeTemplateHelperPlugin, name)

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

	cache.AddToCacheTTLOverride(cache.CacheTypeTemplateHelperPlugin, name, cache.MaxTTL, newPlugin)
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

func pathsFromGroupName(groupName string) ([]string, error) {
	pList := mwsettings.GetSetting("templateHelper/plugins").([]TemplatePluginSettings)
	var outList []string

	for _, plugin := range pList {
		for _, gname := range plugin.Groups {
			if gname == groupName {
				outList = append(outList, plugin.PluginPath)
				break
			}
		}
	}

	return outList, nil
}

func namesFromGroupName(groupName string) ([]string, error) {
	pList := mwsettings.GetSetting("templateHelper/plugins").([]TemplatePluginSettings)
	var outList []string

	for _, plugin := range pList {
		for _, gname := range plugin.Groups {
			if gname == groupName {
				outList = append(outList, plugin.TemplateName)
				break
			}
		}
	}

	return outList, nil
}

func loadPluginFromFile(path string) (iTemplatePlugin, error) {
	newTemplatePlugin := templatePlugin{}
	rawPlugin, pErr := plugin.Open(path)

	if pErr != nil {
		return nil, pErr
	}

	getTemplate, pErr := rawPlugin.Lookup("GetTemplate")
	if pErr != nil {
		return nil, pErr
	}
	getData, pErr := rawPlugin.Lookup("GetData")
	if pErr != nil {
		return nil, pErr
	}

	var bOk bool
	newTemplatePlugin.getTemplateFunc, bOk = getTemplate.(func(t *templateHTML.Template))
	if !bOk {
		return nil, errors.New("GetTemplate() is not of correct format. should be: func(t *templateHTML.Template)")
	}
	newTemplatePlugin.getDataFunc, bOk = getData.(func(argv interface{}) interface{})
	if !bOk {
		return nil, errors.New("GetData() is not of correct format. should be: func(argv interface{}) interface{}")
	}

	return &newTemplatePlugin, nil
}
