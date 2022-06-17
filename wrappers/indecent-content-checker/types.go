package indecent_content_checker

import (
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
	"time"
)

//goland:noinspection GoNameStartsWithPackageName
type IndecentContentCheckerWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

type GetPredictionsRequest struct {
	ImageUrl string `json:"image_url"`
}
type PredictionItem struct {
	ClassName   string  `json:"className"` // "Porn", "Sexy", "Hentai", "Neutral", "Drawing"
	Probability float64 `json:"probability"`
}

type IIndecentContentCheckerWrapper interface {
	GetPredictions(req GetPredictionsRequest, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[[]PredictionItem]
}
