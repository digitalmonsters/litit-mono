package translation

type Language string

const (
	LanguageEn   = Language("en")
	LanguageDe   = Language("de")
	LanguageEs   = Language("es")
	LanguageFrFr = Language("fr-FR")
	LanguageIt   = Language("it")
	LanguagePt   = Language("pt")
)

var SupportedLanguages = map[Language]bool{
	LanguageEn:   true,
	LanguageDe:   true,
	LanguageEs:   true,
	LanguageFrFr: true,
	LanguageIt:   true,
	LanguagePt:   true,
}

const DefaultUserLanguage = LanguageEn

type Place string

const (
	PlaceNotifications = Place("notifications")
)

var SupportedPlaces = map[Place]bool{
	PlaceNotifications: true,
}
