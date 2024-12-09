package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

type Request[T any] struct {
	Params *T

	id string
}

func (r *Request[T]) ID() string {
	return r.id
}

func NewRequest[T any](params *T) *Request[T] {
	return &Request[T]{
		Params: params,
	}
}

type Response[T any] struct {
	Result *T
}

func NewResponse[T any](result *T) *Response[T] {
	return &Response[T]{
		Result: result,
	}
}

type Error struct {
	code int
	err  error
}

func (e *Error) Error() string {
	return e.err.Error()
}

func NewError(code int, underlying error) *Error {
	return &Error{code: code, err: underlying}
}