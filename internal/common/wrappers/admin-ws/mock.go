package admin_ws

import (
	"go.elastic.co/apm"
)

type AdminWsWrapperMock struct {
	SendMessageFn func(event EventType, message interface{}, transaction *apm.Transaction, forceLog bool) chan SendMessageResponseCh
}

func (w *AdminWsWrapperMock) SendMessage(event EventType, message interface{}, transaction *apm.Transaction, forceLog bool) chan SendMessageResponseCh {
	return w.SendMessageFn(event, message, transaction, forceLog)
}

func GetMock() IAdminWsWrapper { // for compiler errors
	return &AdminWsWrapperMock{}
}
