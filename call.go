package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

func call[P any, R any](ctx context.Context, c *base, method string, req *Request[P]) (*Response[R], error) {
	id, inbox := c.router.Add()

	var interceptor Interceptor
	if len(c.interceptors) > 0 {
		interceptor = newStack(c.interceptors)
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

		msgID := json.Number(request.ID())
		msgVersion := "2.0"
		msgParams := json.RawMessage(rawmsg)

		msg := &Message{
			ID:      &msgID,
			JsonRPC: &msgVersion,
			Method:  &method,
			Params:  &msgParams,
		}

		if err := c.stream.Send(msg); err != nil {
			return nil, err
		}

		var result R

		select {
		case resp := <-inbox:
			if resp.Error != nil {
				return nil, NewError(resp.Error.Code, errors.New(resp.Error.Message))
			}
			if resp.Result == nil {
				return nil, fmt.Errorf("no result")
			}
			if err := json.Unmarshal(*resp.Result, &result); err != nil {
				return nil, err
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		return NewResponse(&result), nil
	})

	req.id = strconv.FormatUint(id, 10)
	req.method = method

	resp, err := interceptor.WrapUnary(inner)(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.(*Response[R]), nil
}
