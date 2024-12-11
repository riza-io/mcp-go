package mcp

import (
	"context"
	"encoding/json"
)

func notify[P any](ctx context.Context, c *callable, method string, req *Request[P]) error {
	var interceptor Interceptor
	if len(c.Interceptors) > 0 {
		interceptor = newStack(c.Interceptors)
	} else {
		interceptor = UnaryInterceptorFunc(
			func(next UnaryFunc) UnaryFunc {
				return UnaryFunc(func(ctx context.Context, request AnyRequest) (AnyResponse, error) {
					return next(ctx, request)
				})
			},
		)
	}

	inner := UnaryFunc(func(ctx context.Context, request AnyRequest) (AnyResponse, error) {
		rawmsg, err := json.Marshal(req.Params)
		if err != nil {
			return nil, err
		}

		msgVersion := "2.0"
		msgParams := json.RawMessage(rawmsg)

		msg := &Message{
			JsonRPC: &msgVersion,
			Method:  &method,
			Params:  &msgParams,
		}

		return nil, c.Stream.Send(msg)
	})

	req.method = method

	_, err := interceptor.WrapUnary(inner)(ctx, req)
	if err != nil {
		return err
	}

	return nil
}
