package mcp

import (
	"context"
)

type empty struct{}

func noop[T any](method func(ctx context.Context, req *Request[T])) func(context.Context, *Request[T]) (*Response[empty], error) {
	return func(ctx context.Context, req *Request[T]) (*Response[empty], error) {
		method(ctx, req)
		return NewResponse(&empty{}), nil
	}
}
