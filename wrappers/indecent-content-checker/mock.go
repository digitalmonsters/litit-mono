package indecent_content_checker

import (
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
)

//goland:noinspection ALL
type IndecentContentCheckerWrapperMock struct {
	GetPredictionsFn func(req GetPredictionsRequest, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[[]PredictionItem]
}

func (m *IndecentContentCheckerWrapperMock) GetPredictions(req GetPredictionsRequest, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[[]PredictionItem] {
	return m.GetPredictionsFn(req, apmTransaction, forceLog)
}

func GetMock() IIndecentContentCheckerWrapper { // for compiler errors
	return &IndecentContentCheckerWrapperMock{}
}
