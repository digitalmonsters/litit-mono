package translation

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/translation/translations"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v4"
	"strings"
)

// language -> place -> key -> value
var translationMap = map[Language]map[Place]map[string]string{}

func init() {
	result, err := getTranslationMap()
	if err != nil {
		panic(err)
	}

	translationMap = result
}

func getTranslationMap() (map[Language]map[Place]map[string]string, error) {
	entries, err := translations.Files.ReadDir(".")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := map[Language]map[Place]map[string]string{}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		var fileBytes []byte
		fileBytes, err = translations.Files.ReadFile(fileName)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		var placeMap map[Place]map[string]string
		if err = json.Unmarshal(fileBytes, &placeMap); err != nil {
			return nil, errors.WithStack(err)
		}

		language := Language(strings.Split(fileName, ".json")[0])

		if supportedLanguage := SupportedLanguages[language]; !supportedLanguage {
			continue
		}

		result[language] = placeMap
	}

	return result, nil
}

func GetTranslation(defaultLanguage Language, language Language, place Place, key string) (value null.String, fromRequestedLanguage bool) {
	if supportedDefaultLanguage := SupportedLanguages[defaultLanguage]; !supportedDefaultLanguage {
		return null.String{}, false
	}

	isFromRequestedLanguage := true

	if supportedLanguage := SupportedLanguages[language]; !supportedLanguage {
		language = defaultLanguage
		isFromRequestedLanguage = false
	}

	if supportedPlace := SupportedPlaces[place]; !supportedPlace {
		return null.String{}, isFromRequestedLanguage
	}

	languageMap, ok := translationMap[language]
	if !ok {
		if language == defaultLanguage {
			return null.String{}, isFromRequestedLanguage
		}

		isFromRequestedLanguage = false

		languageMap, ok = translationMap[defaultLanguage]
		if !ok {
			return null.String{}, isFromRequestedLanguage
		}

		language = defaultLanguage
	}

	placeMap, ok := languageMap[place]
	if !ok {
		if language == defaultLanguage {
			return null.String{}, isFromRequestedLanguage
		}

		isFromRequestedLanguage = false

		languageMap, ok = translationMap[defaultLanguage]
		if !ok {
			return null.String{}, isFromRequestedLanguage
		}

		language = defaultLanguage

		placeMap, ok = languageMap[place]
		if !ok {
			return null.String{}, isFromRequestedLanguage
		}
	}

	resultValue, ok := placeMap[key]
	if !ok {
		if language == defaultLanguage {
			return null.String{}, isFromRequestedLanguage
		}

		isFromRequestedLanguage = false

		languageMap, ok = translationMap[defaultLanguage]
		if !ok {
			return null.String{}, isFromRequestedLanguage
		}

		placeMap, ok = languageMap[place]
		if !ok {
			return null.String{}, isFromRequestedLanguage
		}

		resultValue, ok = placeMap[key]
		if !ok {
			return null.String{}, isFromRequestedLanguage
		}
	}

	return null.StringFrom(resultValue), isFromRequestedLanguage
}
