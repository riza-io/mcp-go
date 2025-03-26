package mcp

import (
	"context"
	"fmt"
)

type Method string

const (
	MethodInitialize            Method = "initialize"
	MethodCompletion            Method = "completion/complete"
	MethodListTools             Method = "tools/list"
	MethodCallTool              Method = "tools/call"
	MethodListPrompts           Method = "prompts/list"
	MethodGetPrompt             Method = "prompts/get"
	MethodListResources         Method = "resources/list"
	MethodReadResource          Method = "resources/read"
	MethodListResourceTemplates Method = "resources/templates/list"
	MethodPing                  Method = "ping"
	MethodSetLogLevel           Method = "logging/setLevel"
	MethodNotificationsMessage  Method = "notifications/message"
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

type serverConfig struct {
	interceptors []Interceptor
}

type Server struct {
	handler ServerHandler
	base    *base
}

func NewServer(stream Stream, handler ServerHandler, opts ...Option) *Server {
	cfg := &serverConfig{}
	for _, opt := range opts {
		opt.applyToServer(cfg)
	}
	return &Server{
		handler: handler,
		base: &base{
			router:       newRouter(),
			interceptors: cfg.interceptors,
			stream:       stream,
		},
	}
}

func (s *Server) Listen(ctx context.Context) error {
	return s.base.listen(ctx, s.processMessage)
}

func (s *Server) Ping(ctx context.Context, request *Request[PingRequest]) (*Response[PingResponse], error) {
	return call[PingRequest, PingResponse](ctx, s.base, "ping", request)
}

func (s *Server) LogMessage(ctx context.Context, request *Request[LogMessageRequest]) error {
	return notify[LogMessageRequest](ctx, s.base, "notifications/message", request)
}

func (s *Server) ToolsListChanged(ctx context.Context) error {
	return notify[emptyRequest](ctx, s.base, "notifications/tools/list_changed", NewRequest(&emptyRequest{}))
}

func (s *Server) PromptsListChanged(ctx context.Context) error {
	return notify[emptyRequest](ctx, s.base, "notifications/prompts/list_changed", NewRequest(&emptyRequest{}))
}

func (s *Server) ResourcesListChanged(ctx context.Context) error {
	return notify[emptyRequest](ctx, s.base, "notifications/resources/list_changed", NewRequest(&emptyRequest{}))
}

func (s *Server) processMessage(ctx context.Context, msg *Message) error {
	rr, err := s.ServeMCP(ctx, msg)
	if err != nil {
		return err
	}
	if rr == nil {
		return nil
	}
	return s.base.stream.Send(rr)
}
