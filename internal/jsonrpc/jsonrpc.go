package jsonrpc

import "encoding/json"

type Request struct {
	ID      json.Number     `json:"id"`
	JsonRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type Response[T any] struct {
	ID      json.Number `json:"id"`
	JsonRPC string      `json:"jsonrpc"`
	Result  T           `json:"result,omitempty"`
}

type Error[T any] struct {
	ID      json.Number    `json:"id"`
	JsonRPC string         `json:"jsonrpc"`
	Error   ErrorDetail[T] `json:"error"`
}

type ErrorDetail[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}
