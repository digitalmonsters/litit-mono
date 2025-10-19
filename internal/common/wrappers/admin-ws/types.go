package admin_ws

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/rpc"
)

type EventType string

const (
	WithdrawalTransactionsEventType = EventType("withdrawal_transactions")
	WithdrawalLockEventType         = EventType("withdrawal_lock")
)

func GetSupportedEvents() []EventType {
	return []EventType{WithdrawalTransactionsEventType, WithdrawalLockEventType}
}

type SendMessageRequest struct {
	Event   EventType       `json:"event"`
	Message json.RawMessage `json:"message"`
}

type SendMessageResponseCh struct {
	Error *rpc.RpcError
}
