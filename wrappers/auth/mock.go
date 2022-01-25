package auth

import "go.elastic.co/apm"

type AuthWrapperMock struct {
	ParseTokenFn func(token string, ignoreExpiration bool, apmTransaction *apm.Transaction,
		forceLog bool) chan AuthParseTokenResponseChan
	GenerateTokenFn     func(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GenerateTokenResponseChan
}

func (w *AuthWrapperMock) ParseToken(token string, ignoreExpiration bool, apmTransaction *apm.Transaction,
	forceLog bool) chan AuthParseTokenResponseChan {
	return w.ParseTokenFn(token, ignoreExpiration, apmTransaction, forceLog)
}

func (w *AuthWrapperMock) GenerateToken(userId int64, apmTransaction *apm.Transaction,
	forceLog bool) chan GenerateTokenResponseChan {
	return w.GenerateTokenFn(userId, apmTransaction, forceLog)
}

func GetMock() IAuthWrapper { // for compiler errors
	return &AuthWrapperMock{}
}
