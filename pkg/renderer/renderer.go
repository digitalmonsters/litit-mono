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

func Render(renderTemplate database.RenderTemplate, renderingData map[string]string, language translation.Language) (title string,
	body string, headline string, titleMultiple string, bodyMultiple string, headlineMultiple string, err error) {
	prefix := fmt.Sprintf("%v_%v", renderTemplate.Id, renderTemplate.UpdatedAt)

	translatedTitle, _ := translation.GetTranslation(translation.DefaultUserLanguage, language, translation.PlaceNotifications, fmt.Sprintf("%v_title", renderTemplate.Id))
	title, err = RenderText(fmt.Sprintf("%v_title", prefix), translatedTitle.ValueOrZero(), renderingData)

	if err != nil {
		return "", "", "", "", "", "", err
	}

	translatedBody, _ := translation.GetTranslation(translation.DefaultUserLanguage, language, translation.PlaceNotifications, fmt.Sprintf("%v_body", renderTemplate.Id))
	body, err = RenderText(fmt.Sprintf("%v_body", prefix), translatedBody.ValueOrZero(), renderingData)

	if err != nil {
		return "", "", "", "", "", "", err
	}

	translatedHeadline, _ := translation.GetTranslation(translation.DefaultUserLanguage, language, translation.PlaceNotifications, fmt.Sprintf("%v_headline", renderTemplate.Id))
	headline, err = RenderText(fmt.Sprintf("%v_headline", prefix), translatedHeadline.ValueOrZero(), renderingData)

	if err != nil {
		return "", "", "", "", "", "", err
	}

	translatedTitleMultiple, _ := translation.GetTranslation(translation.DefaultUserLanguage, language, translation.PlaceNotifications, fmt.Sprintf("%v_title_multiple", renderTemplate.Id))
	titleMultiple, err = RenderText(fmt.Sprintf("%v_title_multiple", prefix), translatedTitleMultiple.ValueOrZero(), renderingData)

	if err != nil {
		return "", "", "", "", "", "", err
	}

	translatedBodyMultiple, _ := translation.GetTranslation(translation.DefaultUserLanguage, language, translation.PlaceNotifications, fmt.Sprintf("%v_body_multiple", renderTemplate.Id))
	bodyMultiple, err = RenderText(fmt.Sprintf("%v_body_multiple", prefix), translatedBodyMultiple.ValueOrZero(), renderingData)

	if err != nil {
		return "", "", "", "", "", "", err
	}

	translatedHeadlineMultiple, _ := translation.GetTranslation(translation.DefaultUserLanguage, language, translation.PlaceNotifications, fmt.Sprintf("%v_headline_multiple", renderTemplate.Id))
	headlineMultiple, err = RenderText(fmt.Sprintf("%v_headline_multiple", prefix), translatedHeadlineMultiple.ValueOrZero(), renderingData)

	if err != nil {
		return "", "", "", "", "", "", err
	}

	return title, body, headline, titleMultiple, bodyMultiple, headlineMultiple, err
}

func RenderText(templateName string, templateBody string, renderingData map[string]string) (string, error) {
	if len(templateBody) == 0 {
		return "", nil
	}

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
