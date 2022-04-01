package auth

import (
	"context"
	"go.elastic.co/apm"
)

type AuthWrapperMock struct {
	ParseTokenFn func(token string, ignoreExpiration bool, apmTransaction *apm.Transaction,
		forceLog bool) chan AuthParseTokenResponseChan
	ParseNewAdminTokenFn func(token string, ignoreExpiration bool, apmTransaction *apm.Transaction,
		forceLog bool) chan AuthParseTokenResponseChan
	GenerateTokenFn         func(userId int64, isGuest bool, meta MetaData, apmTransaction *apm.Transaction, forceLog bool) chan GenerateTokenResponseChan
	GenerateNewAdminTokenFn func(userId int64, ctx context.Context, forceLog bool) chan GenerateTokenResponseChan
}

func (w *AuthWrapperMock) ParseToken(token string, ignoreExpiration bool, apmTransaction *apm.Transaction,
	forceLog bool) chan AuthParseTokenResponseChan {
	return w.ParseTokenFn(token, ignoreExpiration, apmTransaction, forceLog)
}

func (w *AuthWrapperMock) ParseNewAdminToken(token string, ignoreExpiration bool, apmTransaction *apm.Transaction,
	forceLog bool) chan AuthParseTokenResponseChan {
	return w.ParseNewAdminTokenFn(token, ignoreExpiration, apmTransaction, forceLog)
}

func (w *AuthWrapperMock) GenerateToken(userId int64, isGuest bool, meta MetaData, apmTransaction *apm.Transaction,
	forceLog bool) chan GenerateTokenResponseChan {
	return w.GenerateTokenFn(userId, isGuest, meta, apmTransaction, forceLog)
}

func (w *AuthWrapperMock) GenerateNewAdminToken(userId int64, ctx context.Context,
	forceLog bool) chan GenerateTokenResponseChan {
	return w.GenerateNewAdminTokenFn(userId, ctx, forceLog)
}

func GetMock() IAuthWrapper { // for compiler errors
	return &AuthWrapperMock{}
}
