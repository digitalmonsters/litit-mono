package notification_gateway

import (
	"github.com/digitalmonsters/go-common/rpc"
)

type SendMessageRequest struct {
	Message     string `json:"message"`
	PhoneNumber string `json:"phone_number"`
}

type SendMessageResponseChan struct {
	Error *rpc.RpcError `json:"error"`
}
