package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

func (s *Server) ServeMCP(ctx context.Context, msg *Message) (*Message, error) {
	h := s.handler
	switch m := Method(*msg.Method); m {
	case MethodInitialize:
		return serveMCP(ctx, s.base, msg, h.Initialize)
	case MethodCompletion:
		return serveMCP(ctx, s.base, msg, h.Completion)
	case MethodListTools:
		return serveMCP(ctx, s.base, msg, h.ListTools)
	case MethodCallTool:
		return serveMCP(ctx, s.base, msg, h.CallTool)
	case MethodListPrompts:
		return serveMCP(ctx, s.base, msg, h.ListPrompts)
	case MethodGetPrompt:
		return serveMCP(ctx, s.base, msg, h.GetPrompt)
	case MethodListResources:
		return serveMCP(ctx, s.base, msg, h.ListResources)
	case MethodReadResource:
		return serveMCP(ctx, s.base, msg, h.ReadResource)
	case MethodListResourceTemplates:
		return serveMCP(ctx, s.base, msg, h.ListResourceTemplates)
	case MethodPing:
		return serveMCP(ctx, s.base, msg, h.Ping)
	case MethodSetLogLevel:
		return serveMCP(ctx, s.base, msg, h.SetLogLevel)
	default:
		return nil, fmt.Errorf("unknown method: %s", m)
	}
}

func serveMCP[T, V any](ctx context.Context, cfg *base, msg *Message, method func(ctx context.Context, req *Request[T]) (*Response[V], error)) (*Message, error) {
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

	if msg.Params != nil && len(*msg.Params) > 0 {
		if err := json.Unmarshal(*msg.Params, &params); err != nil {
			return nil, err
		}
	}

	req := NewRequest(&params)
	if msg.ID != nil {
		req.id = msg.ID.String()
	}
	if msg.Metadata != nil {
		req.metadata = msg.Metadata
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
		return nil, nil
	}

	if err != nil {
		return &Message{
			Metadata: msg.Metadata,
			ID:       msg.ID,
			JsonRPC:  msg.JsonRPC,
			Error: &ErrorDetail{
				Code:    9,
				Message: err.Error(),
			},
		}, nil
	}

	resp := rr.(*Response[V])

	rawresult, err := json.Marshal(resp.Result)
	if err != nil {
		return &Message{
			Metadata: msg.Metadata,
			ID:       msg.ID,
			JsonRPC:  msg.JsonRPC,
			Error: &ErrorDetail{
				Code:    9,
				Message: err.Error(),
			},
		}, nil
	}

	rawmsg := json.RawMessage(rawresult)
	return &Message{
		Metadata: msg.Metadata,
		ID:       msg.ID,
		JsonRPC:  msg.JsonRPC,
		Result:   &rawmsg,
	}, nil
}
