package templateHelper

import (
	templateHTML "html/template"
	"io"
	"reflect"
	templateText "text/template"

	"github.com/CanadianCommander/MicroWeb/pkg/logger"
	mwsettings "github.com/CanadianCommander/MicroWeb/pkg/mwSettings"
)

/*
AddTemplate gets the Template structure with the given template name (as set in configuration file).
*/
func AddTemplate(t *templateHTML.Template, name string) (*templateHTML.Template, error) {
	plug, err := getPlugin(name)
	if err != nil {
		logger.LogError("Failed to get plugin for template: %s with error: %s", name, err.Error())
		return nil, err
	}

	plug.getTemplate(t.New(name))
	t.Funcs(map[string]interface{}{name: plug.getData})
	return t, nil
}

/*
ProcessTemplateHTML takes the template described by templateFileBuffer and uses the html/template
package to parse and execute the template, pushing output on the, out io.Writer.
The big difference between this and ProcessTemplateText, is that this function performs HTML escaping of text.
*/
func ProcessTemplateHTML(templateFileBuffer *[]byte, out io.Writer, tStruct interface{}) error {
	templateParser := templateHTML.New("root")
	_, tErr := templateParser.Parse(string((*templateFileBuffer)[:]))
	if tErr != nil {
		logger.LogError("could not parse template file w/ error: %s", tErr.Error())
		return tErr
	}

	return templateParser.Execute(out, tStruct)
}

/*
ProcessTemplateText takes the template described by templateFileBuffer and uses the text/template
package to parse and execute the template, pushing output on the, out io.Writer.
*/
func ProcessTemplateText(templateFileBuffer *[]byte, out io.Writer, tStruct interface{}) error {
	templateParser := templateText.New("root")
	_, tErr := templateParser.Parse(string((*templateFileBuffer)[:]))
	if tErr != nil {
		logger.LogError("could not parse template file w/ error: %s", tErr.Error())
		return tErr
	}

	return templateParser.Execute(out, tStruct)
}

//TemplatePluginSettings represents the settings for a template plugin
type TemplatePluginSettings struct {
	PluginPath,
	TemplateName string
}

// AddTemplateHelperSettingDecoders adds a decoder for the template Helper setting format in the config file.
func AddTemplateHelperSettingDecoders() {
	var templateHelperPath = "templateHelper/plugins"

	mwsettings.AddSettingDecoder(mwsettings.NewFunctionalSettingDecoder(func(s interface{}) (string, interface{}) {
		if reflect.ValueOf(s).Type().Kind() == reflect.Slice {
			pList := s.([]interface{})
			outList := make([]TemplatePluginSettings, len(pList))

			for i, p := range pList {
				outList[i] = TemplatePluginSettings{}
				outList[i].PluginPath = p.(map[string]interface{})["plugin"].(string)
				outList[i].TemplateName = p.(map[string]interface{})["name"].(string)
			}
			return templateHelperPath, outList
		}

		logger.LogError("Error parsing templateHelper plugin list. format incorrect")
		return "ERROR", nil
	},
		func(path string) bool {
			if path == templateHelperPath {
				return true
			}
			return false
		}))
}
