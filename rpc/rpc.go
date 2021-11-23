package rpc

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/error_codes"
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
	Code    error_codes.ErrorCode  `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
	Stack   string                 `json:"stack"`
}
