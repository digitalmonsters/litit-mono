package nodejs

import "encoding/json"

type Response struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *ResponseError  `json:"error"`
}

type ResponseError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}
