package mcp

import "encoding/json"

type Stream interface {
	Recv() (*Message, error)
	Send(msg *Message) error
}

type Message struct {
	ID      *json.Number     `json:"id,omitempty"`
	JsonRPC *string          `json:"jsonrpc"`
	Method  *string          `json:"method,omitempty"`
	Params  *json.RawMessage `json:"params,omitempty"`
	Result  *json.RawMessage `json:"result,omitempty"`
	Error   *ErrorDetail     `json:"error,omitempty"`

	// Metadata is used to store additional information about the message for
	// processing by the client or server. It is never sent across the wire.
	Metadata map[string]string `json:"-"`
}

type ErrorDetail struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}
