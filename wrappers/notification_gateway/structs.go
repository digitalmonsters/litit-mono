package notification_gateway

import (
	"github.com/digitalmonsters/go-common/rpc"
)

type SendSmsMessageRequest struct {
	Message     string `json:"message"`
	PhoneNumber string `json:"phone_number"`
}

type SendSmsMessageResponseChan struct {
	Error *rpc.RpcError `json:"error"`
}

type SendEmailMessageRequest struct {
	CcAddresses []string `json:"cc_addresses"`
	ToAddresses []string `json:"to_addresses"`
	HtmlBody    string   `json:"html_body"`
	TextBody    string   `json:"text_body"`
	Subject     string   `json:"subject"`
}

type SendEmailMessageResponseChan struct {
	Error *rpc.RpcError `json:"error"`
}
