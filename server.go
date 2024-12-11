package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

type ServerHandler interface {
	Initialize(ctx context.Context, req *Request[InitializeRequest]) (*Response[InitializeResponse], error)
	ListTools(ctx context.Context, req *Request[ListToolsRequest]) (*Response[ListToolsResponse], error)
	CallTool(ctx context.Context, req *Request[CallToolRequest]) (*Response[CallToolResponse], error)
	ListPrompts(ctx context.Context, req *Request[ListPromptsRequest]) (*Response[ListPromptsResponse], error)
	GetPrompt(ctx context.Context, req *Request[GetPromptRequest]) (*Response[GetPromptResponse], error)
	ListResources(ctx context.Context, req *Request[ListResourcesRequest]) (*Response[ListResourcesResponse], error)
	ReadResource(ctx context.Context, req *Request[ReadResourceRequest]) (*Response[ReadResourceResponse], error)
	ListResourceTemplates(ctx context.Context, req *Request[ListResourceTemplatesRequest]) (*Response[ListResourceTemplatesResponse], error)
	Completion(ctx context.Context, req *Request[CompletionRequest]) (*Response[CompletionResponse], error)
	Ping(ctx context.Context, req *Request[PingRequest]) (*Response[PingResponse], error)
	SetLogLevel(ctx context.Context, req *Request[SetLogLevelRequest]) (*Response[SetLogLevelResponse], error)
}

type UnimplementedServer struct{}

func (s *UnimplementedServer) Initialize(ctx context.Context, req *Request[InitializeRequest]) (*Response[InitializeResponse], error) {
	return nil, fmt.Errorf("unimplemented")
}

func (s *UnimplementedServer) ListTools(ctx context.Context, req *Request[ListToolsRequest]) (*Response[ListToolsResponse], error) {
	return nil, fmt.Errorf("unimplemented")
}

func (s *UnimplementedServer) CallTool(ctx context.Context, req *Request[CallToolRequest]) (*Response[CallToolResponse], error) {
	return nil, fmt.Errorf("unimplemented")
}

func (s *UnimplementedServer) ListPrompts(ctx context.Context, req *Request[ListPromptsRequest]) (*Response[ListPromptsResponse], error) {
	return nil, fmt.Errorf("unimplemented")
}

func (s *UnimplementedServer) GetPrompt(ctx context.Context, req *Request[GetPromptRequest]) (*Response[GetPromptResponse], error) {
	return nil, fmt.Errorf("unimplemented")
}

func (s *UnimplementedServer) ListResources(ctx context.Context, req *Request[ListResourcesRequest]) (*Response[ListResourcesResponse], error) {
	return nil, fmt.Errorf("unimplemented")
}

func (s *UnimplementedServer) ReadResource(ctx context.Context, req *Request[ReadResourceRequest]) (*Response[ReadResourceResponse], error) {
	return nil, fmt.Errorf("unimplemented")
}

func (s *UnimplementedServer) ListResourceTemplates(ctx context.Context, req *Request[ListResourceTemplatesRequest]) (*Response[ListResourceTemplatesResponse], error) {
	return nil, fmt.Errorf("unimplemented")
}

func (s *UnimplementedServer) Completion(ctx context.Context, req *Request[CompletionRequest]) (*Response[CompletionResponse], error) {
	return nil, fmt.Errorf("unimplemented")
}

func (s *UnimplementedServer) Ping(ctx context.Context, req *Request[PingRequest]) (*Response[PingResponse], error) {
	return NewResponse(&PingResponse{}), nil
}

func (s *UnimplementedServer) SetLogLevel(ctx context.Context, req *Request[SetLogLevelRequest]) (*Response[SetLogLevelResponse], error) {
	return nil, fmt.Errorf("unimplemented")
}

func process[T, V any](ctx context.Context, cfg *callable, msg *Message, method func(ctx context.Context, req *Request[T]) (*Response[V], error)) error {
	var interceptor Interceptor
	if len(cfg.Interceptors) > 0 {
		interceptor = newStack(cfg.Interceptors)
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
			return err
		}
	}

	req := NewRequest(&params)
	req.id = msg.ID.String()
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
	if err != nil {
		return cfg.Stream.Send(&Message{
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
		return cfg.Stream.Send(&Message{
			ID:      msg.ID,
			JsonRPC: msg.JsonRPC,
			Error: &ErrorDetail{
				Code:    9,
				Message: err.Error(),
			},
		})
	}

	rawmsg := json.RawMessage(rawresult)
	return cfg.Stream.Send(&Message{
		ID:      msg.ID,
		JsonRPC: msg.JsonRPC,
		Result:  &rawmsg,
	})
}

type serverConfig struct {
	interceptors []Interceptor
}

type Server struct {
	cfg    *serverConfig
	srv    ServerHandler
	stream Stream
}

func NewServer(stream Stream, handler ServerHandler, opts ...Option) *Server {
	cfg := &serverConfig{}
	for _, opt := range opts {
		opt.applyToServer(cfg)
	}
	return &Server{
		cfg:    cfg,
		stream: stream,
		srv:    handler,
	}
}

func (s Server) Listen(ctx context.Context) error {
	for {
		msg, err := s.stream.Recv()

		if err != nil {
			return err
		}

		go func() {
			s.processMessage(ctx, msg)
		}()
	}
}

func (s Server) processMessage(ctx context.Context, msg *Message) error {
	if msg.Method == nil {
		return nil
	}

	cfg := &callable{
		Interceptors: s.cfg.interceptors,
		Stream:       s.stream,
	}
	srv := s.srv

	switch m := *msg.Method; m {
	case "initialize":
		return process(ctx, cfg, msg, srv.Initialize)
	case "completion/complete":
		return process(ctx, cfg, msg, srv.Completion)
	case "tools/list":
		return process(ctx, cfg, msg, srv.ListTools)
	case "tools/call":
		return process(ctx, cfg, msg, srv.CallTool)
	case "prompts/list":
		return process(ctx, cfg, msg, srv.ListPrompts)
	case "prompts/get":
		return process(ctx, cfg, msg, srv.GetPrompt)
	case "resources/list":
		return process(ctx, cfg, msg, srv.ListResources)
	case "resources/read":
		return process(ctx, cfg, msg, srv.ReadResource)
	case "resources/templates/list":
		return process(ctx, cfg, msg, srv.ListResourceTemplates)
	case "ping":
		return process(ctx, cfg, msg, srv.Ping)
	case "logging/setLevel":
		return process(ctx, cfg, msg, srv.SetLogLevel)
	default:
		return fmt.Errorf("unknown method: %s", m)
	}

}
