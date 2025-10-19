package base_api

import "go.elastic.co/apm"

//goland:noinspection ALL
type BaseApiWrapperMock struct {
	GetCountriesWithAgeLimitFn func(apmTransaction *apm.Transaction, forceLog bool) chan GetCountriesWithAgeLimitResponseChan
}

func (m *BaseApiWrapperMock) GetCountriesWithAgeLimit(apmTransaction *apm.Transaction, forceLog bool) chan GetCountriesWithAgeLimitResponseChan {
	return m.GetCountriesWithAgeLimitFn(apmTransaction, forceLog)
}

func GetMock() IBaseApiWrapper { // for compiler errors
	return &BaseApiWrapperMock{}
}

