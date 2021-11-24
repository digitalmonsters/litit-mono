package auth

import "go.elastic.co/apm"

type AuthWrapperMock struct {
	ParseTokenFn func(token string, ignoreExpiration bool, apmTransaction *apm.Transaction,
		forceLog bool) chan AuthParseTokenResponseChan
}

func (w *AuthWrapperMock) ParseToken(token string, ignoreExpiration bool, apmTransaction *apm.Transaction,
	forceLog bool) chan AuthParseTokenResponseChan {
	return w.ParseTokenFn(token, ignoreExpiration, apmTransaction, forceLog)
}

func GetMock() IAuthWrapper { // for compiler errors
	return &AuthWrapperMock{}
}
