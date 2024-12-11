package mcp

import (
	"context"
	"fmt"
)

type ClientHandler interface {
	Sampling(ctx context.Context, request *Request[SamplingRequest]) (*Response[SamplingResponse], error)
	Ping(ctx context.Context, request *Request[PingRequest]) (*Response[PingResponse], error)
	LogMessage(ctx context.Context, request *Request[LogMessageRequest])
}

type UnimplementedClient struct{}

func (u *UnimplementedClient) Sampling(ctx context.Context, request *Request[SamplingRequest]) (*Response[SamplingResponse], error) {
	return nil, fmt.Errorf("not implemented")
}

func (u *UnimplementedClient) LogMessage(ctx context.Context, request *Request[LogMessageRequest]) {
}

func (c *UnimplementedClient) Ping(ctx context.Context, req *Request[PingRequest]) (*Response[PingResponse], error) {
	return NewResponse(&PingResponse{}), nil
}

type Client struct {
	handler      ClientHandler
	interceptors []Interceptor
	base         *base
}

func NewClient(stream Stream, handler ClientHandler, opts ...Option) *Client {
	c := &Client{
		handler: handler,
	}
	for _, opt := range opts {
		opt.applyToClient(c)
	}
	c.base = &base{
		router:       newRouter(),
		interceptors: c.interceptors,
		stream:       stream,
	}
	return c
}

// sync.Once?
func (c *Client) Listen(ctx context.Context) error {
	return c.base.listen(ctx, c.processMessage)
}

func (c *Client) processMessage(ctx context.Context, msg *Message) error {
	srv := c.handler
	switch m := *msg.Method; m {
	case "ping":
		return process(ctx, c.base, msg, srv.Ping)
	case "notifications/message":
		return process(ctx, c.base, msg, noop(srv.LogMessage))
	default:
		return fmt.Errorf("unknown method: %s", m)
	}
}

func (c *Client) Initialize(ctx context.Context, request *Request[InitializeRequest]) (*Response[InitializeResponse], error) {
	return call[InitializeRequest, InitializeResponse](ctx, c.base, "initialize", request)
}

func (c *Client) ListResources(ctx context.Context, request *Request[ListResourcesRequest]) (*Response[ListResourcesResponse], error) {
	return call[ListResourcesRequest, ListResourcesResponse](ctx, c.base, "resources/list", request)
}

func (c *Client) ListTools(ctx context.Context, request *Request[ListToolsRequest]) (*Response[ListToolsResponse], error) {
	return call[ListToolsRequest, ListToolsResponse](ctx, c.base, "tools/list", request)
}

func (c *Client) CallTool(ctx context.Context, request *Request[CallToolRequest]) (*Response[CallToolResponse], error) {
	return call[CallToolRequest, CallToolResponse](ctx, c.base, "tools/call", request)
}

func (c *Client) ListPrompts(ctx context.Context, request *Request[ListPromptsRequest]) (*Response[ListPromptsResponse], error) {
	return call[ListPromptsRequest, ListPromptsResponse](ctx, c.base, "prompts/list", request)
}

func (c *Client) GetPrompt(ctx context.Context, request *Request[GetPromptRequest]) (*Response[GetPromptResponse], error) {
	return call[GetPromptRequest, GetPromptResponse](ctx, c.base, "prompts/get", request)
}

func (c *Client) ReadResource(ctx context.Context, request *Request[ReadResourceRequest]) (*Response[ReadResourceResponse], error) {
	return call[ReadResourceRequest, ReadResourceResponse](ctx, c.base, "resources/read", request)
}

func (c *Client) ListResourceTemplates(ctx context.Context, request *Request[ListResourceTemplatesRequest]) (*Response[ListResourceTemplatesResponse], error) {
	return call[ListResourceTemplatesRequest, ListResourceTemplatesResponse](ctx, c.base, "resources/templates/list", request)
}

func (c *Client) Completion(ctx context.Context, request *Request[CompletionRequest]) (*Response[CompletionResponse], error) {
	return call[CompletionRequest, CompletionResponse](ctx, c.base, "completion", request)
}

func (c *Client) Ping(ctx context.Context, request *Request[PingRequest]) (*Response[PingResponse], error) {
	return call[PingRequest, PingResponse](ctx, c.base, "ping", request)
}

func (c *Client) SetLogLevel(ctx context.Context, request *Request[SetLogLevelRequest]) (*Response[SetLogLevelResponse], error) {
	return call[SetLogLevelRequest, SetLogLevelResponse](ctx, c.base, "logging/setLevel", request)
}
