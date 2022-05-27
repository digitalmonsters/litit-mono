package translation

type Language string

const (
	LanguageEn = Language("en")
)

var SupportedLanguages = map[Language]bool{
	LanguageEn: true,
}

const DefaultUserLanguage = LanguageEn

type Place string

const (
	PlaceNotifications = Place("notifications")
)

var SupportedPlaces = map[Place]bool{
	PlaceNotifications: true,
}
