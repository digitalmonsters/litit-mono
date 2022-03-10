package renderer

import (
	"bytes"
	"fmt"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"text/template"
	"time"
)

var templateCache = cache.New(30*time.Minute, 11*time.Minute)
var TemplateRenderingError = errors.New("template rendering error")

func Render(renderTemplate database.RenderTemplate, renderingData map[string]string) (title string, body string, err error) {
	prefix := fmt.Sprintf("%v_%v", renderTemplate.Id, renderTemplate.UpdatedAt)

	title, err = RenderText(fmt.Sprintf("%v_title", prefix), renderTemplate.Title, renderingData)

	if err != nil {
		return "", "", err
	}

	body, err = RenderText(fmt.Sprintf("%v_body", prefix), renderTemplate.Body, renderingData)

	if err != nil {
		return "", "", err
	}

	return title, body, err
}

func RenderText(templateName string, templateBody string, renderingData map[string]string) (string, error) {
	cachedObj, ok := templateCache.Get(templateName)

	var renderingTemplate *template.Template

	if !ok {
		t := template.New(templateName)

		updatedTemplate, err := t.Parse(templateBody)

		if err != nil {
			return "", errors.WithStack(err)
		}

		updatedTemplate = updatedTemplate.Option("missingkey=error")

		templateCache.SetDefault(templateName, updatedTemplate)

		renderingTemplate = updatedTemplate
	} else {
		renderingTemplate = cachedObj.(*template.Template)
	}

	var buffer bytes.Buffer

	if err := renderingTemplate.Execute(&buffer, renderingData); err != nil {
		return buffer.String(), errors.Wrap(TemplateRenderingError, err.Error())
	}

	return buffer.String(), nil
}
