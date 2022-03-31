package notification_gateway

import (
	"fmt"
	"github.com/digitalmonsters/go-common/common"
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
	CcAddresses  []string `json:"cc_addresses"`
	ToAddresses  []string `json:"to_addresses"`
	HtmlBody     string   `json:"html_body"`
	TextBody     string   `json:"text_body"`
	Subject      string   `json:"subject"`
	Template     string   `json:"template"`
	TemplateData string   `json:"template_data"`
	PublishKey   string   `json:"publish_key"`
}

func (v SendEmailMessageRequest) GetPublishKey() string {
	return fmt.Sprintf("%v", v.PublishKey)
}

type SendEmailMessageResponseChan struct {
	Error *rpc.RpcError `json:"error"`
}

type SendPushRequest struct {
	Tokens     []string               `json:"tokens"`
	DeviceType common.DeviceType      `json:"device_type"`
	Title      string                 `json:"title"`
	Body       string                 `json:"body"`
	ExtraData  map[string]interface{} `json:"extra_data"`
	PublishKey string                 `json:"publish_key"`
}

func (v SendPushRequest) GetPublishKey() string {
	return fmt.Sprintf("%v", v.PublishKey)
}
