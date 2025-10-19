package tesseract_ocr_api

type RecognizeImageTextRequest struct {
	ImageData []byte     `json:"image_data"`
	Languages []Language `json:"languages"`
	HocrMode  bool       `json:"hocr_mode"`
	Trim      bool       `json:"trim"`
}

type RecognizeImageTextResponse struct {
	Text string `json:"text"`
}

type Language string

const (
	LanguageEng = Language("eng")
	LanguageSpa = Language("spa")
	LanguageHin = Language("hin")
	LanguageBen = Language("ben")
	LanguagePor = Language("por")
	LanguageRus = Language("rus")
	LanguageJpn = Language("jpn")
	LanguagePan = Language("pan")
	LanguageMar = Language("mar")
	LanguageTel = Language("tel")
	LanguageTur = Language("tur")
	LanguageKor = Language("kor")
	LanguageFra = Language("fra")
)

func GetSupportedLanguages() []Language {
	return []Language{LanguageEng, LanguageSpa, LanguageHin, LanguageBen, LanguagePor, LanguageRus, LanguageJpn,
		LanguagePan, LanguageMar, LanguageTel, LanguageTur, LanguageKor, LanguageFra}
}
