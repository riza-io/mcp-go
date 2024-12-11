package mcp

import (
	"context"
	"encoding/json"
)

type empty struct{}

func noop[T any](method func(ctx context.Context, req *Request[T])) func(context.Context, *Request[T]) (*Response[empty], error) {
	return func(ctx context.Context, req *Request[T]) (*Response[empty], error) {
		method(ctx, req)
		return NewResponse(&empty{}), nil
	}
}

func process[T, V any](ctx context.Context, cfg *base, msg *Message, method func(ctx context.Context, req *Request[T]) (*Response[V], error)) error {
	var interceptor Interceptor
	if len(cfg.interceptors) > 0 {
		interceptor = newStack(cfg.interceptors)
	} else {
		interceptor = UnaryInterceptorFunc(
			func(next UnaryFunc) UnaryFunc {
				return UnaryFunc(func(ctx context.Context, request AnyRequest) (AnyResponse, error) {
					return next(ctx, request)
				})
			},
		)
	}

	var params T
	if err := json.Unmarshal(*msg.Params, &params); err != nil {
		return err
	}
	req := NewRequest(&params)
	if msg.ID != nil {
		req.id = msg.ID.String()
	}
	req.method = *msg.Method

	inner := UnaryFunc(func(ctx context.Context, request AnyRequest) (AnyResponse, error) {
		req := request.(*Request[T])
		resp, rerr := method(ctx, req)
		if rerr != nil {
			return nil, rerr
		}
		resp.id = req.id
		return resp, nil
	})

	rr, err := interceptor.WrapUnary(inner)(ctx, req)

	// If the incoming message has no ID, we don't need to send a response
	if msg.ID == nil {
		return nil
	}

	if err != nil {
		return cfg.stream.Send(&Message{
			ID:      msg.ID,
			JsonRPC: msg.JsonRPC,
			Error: &ErrorDetail{
				Code:    9,
				Message: err.Error(),
			},
		})
	}

	resp := rr.(*Response[V])

	rawresult, err := json.Marshal(resp.Result)
	if err != nil {
		return cfg.stream.Send(&Message{
			ID:      msg.ID,
			JsonRPC: msg.JsonRPC,
			Error: &ErrorDetail{
				Code:    9,
				Message: err.Error(),
			},
		})
	}

	rawmsg := json.RawMessage(rawresult)
	return cfg.stream.Send(&Message{
		ID:      msg.ID,
		JsonRPC: msg.JsonRPC,
		Result:  &rawmsg,
	})
}
