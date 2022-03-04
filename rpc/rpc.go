package rpc

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/pkg/errors"
	"strings"
)

//goland:noinspection ALL
type RpcRequest struct {
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	Id      string          `json:"id"`
	JsonRpc string          `json:"jsonrpc"`
}

//goland:noinspection ALL
type RpcResponse struct {
	JsonRpc           string      `json:"jsonrpc"`
	Result            interface{} `json:"result"`
	Error             *RpcError   `json:"error"`
	Id                string      `json:"id"`
	ExecutionTimingMs int64       `json:"execution_timing"`
	TotalTimingMs     int64       `json:"total_timing_ms"`
	Hostname          string      `json:"hostname"`
}

//goland:noinspection ALL
type RpcRequestInternal struct {
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Id      string      `json:"id"`
	JsonRpc string      `json:"jsonrpc"`
}

//goland:noinspection ALL
type RpcResponseInternal struct {
	JsonRpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *RpcError       `json:"error"`
	Id      string          `json:"id"`
}

//goland:noinspection ALL
type RpcError struct {
	Code        error_codes.ErrorCode  `json:"code"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data"`
	Stack       string                 `json:"stack"`
	Hostname    string                 `json:"hostname"`
	ServiceName string                 `json:"-"`
}

func (r *RpcError) ToError() error {
	var builder strings.Builder

	builder.WriteString("service [")

	if len(r.ServiceName) == 0 {
		builder.WriteString(r.ServiceName)
	} else {
		builder.WriteString(r.Hostname)
	}

	builder.WriteString(fmt.Sprintf("] replied with code [%v] and message [%v]", int(r.Code), r.Message))

	return errors.New(builder.String())
}

type CustomFile struct {
	Data                         []byte `json:"data"`
	Filename                     string `json:"filename"`
	MimeType                     string `json:"mime_type"`
	ContentDispositionFirstParam string `json:"content_disposition_first_param"`
}
