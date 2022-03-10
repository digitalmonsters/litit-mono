package notification_gateway

import (
	"fmt"
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

type DeviceType int

const (
	DeviceTypeIos     = DeviceType(1)
	DeviceTypeAndroid = DeviceType(2)
)

func (d DeviceType) ToString() string {
	switch d {
	case DeviceTypeIos:
		return "ios"
	case DeviceTypeAndroid:
		return "android"
	default:
		return fmt.Sprintf("unk %v", d)
	}
}

type SendPushRequest struct {
	Tokens     []string          `json:"tokens"`
	DeviceType DeviceType        `json:"device_type"`
	Title      string            `json:"title"`
	Body       string            `json:"body"`
	ExtraData  map[string]string `json:"extra_data"`
	PublishKey string            `json:"publish_key"`
}

func (v SendPushRequest) GetPublishKey() string {
	return fmt.Sprintf("%v", v.PublishKey)
}
