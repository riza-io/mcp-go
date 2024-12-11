package mcp

import "encoding/json"

type Stream interface {
	Recv() (*Message, error)
	Send(msg *Message) error
}

type Message struct {
	ID      *json.Number     `json:"id"`
	JsonRPC *string          `json:"jsonrpc"`
	Method  *string          `json:"method"`
	Params  *json.RawMessage `json:"params"`
	Result  *json.RawMessage `json:"result,omitempty"`
	Error   *ErrorDetail     `json:"error,omitempty"`
}

type ErrorDetail struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}
