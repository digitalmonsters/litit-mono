package renderer

import (
	"bytes"
	"fmt"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"text/template"
	"time"
)

var templateCache = cache.New(30*time.Minute, 11*time.Minute)
var TemplateRenderingError = errors.New("template rendering error")

func Render(renderTemplate database.RenderTemplate, renderingData map[string]string, language translation.Language) (title string, body string, headline string, err error) {
	prefix := fmt.Sprintf("%v_%v", renderTemplate.Id, renderTemplate.UpdatedAt)

	if len(renderTemplate.Title) > 0 {
		translatedTitle, _ := translation.GetTranslation(translation.DefaultUserLanguage, language, translation.PlaceNotifications, renderTemplate.Title)
		title, err = RenderText(fmt.Sprintf("%v_title", prefix), translatedTitle.ValueOrZero(), renderingData)

		if err != nil {
			return "", "", "", err
		}
	}

	if len(renderTemplate.Body) > 0 {
		translatedBody, _ := translation.GetTranslation(translation.DefaultUserLanguage, language, translation.PlaceNotifications, renderTemplate.Body)
		body, err = RenderText(fmt.Sprintf("%v_body", prefix), translatedBody.ValueOrZero(), renderingData)

		if err != nil {
			return "", "", "", err
		}
	}

	if len(renderTemplate.Headline) > 0 {
		translatedHeadline, _ := translation.GetTranslation(translation.DefaultUserLanguage, language, translation.PlaceNotifications, renderTemplate.Headline)
		headline, err = RenderText(fmt.Sprintf("%v_headline", prefix), translatedHeadline.ValueOrZero(), renderingData)

		if err != nil {
			return "", "", "", err
		}
	}

	return title, body, headline, err
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
