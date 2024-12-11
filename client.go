package mcp

import (
	"context"
	"fmt"
	"strconv"
)

type ClientHandler interface {
	Sampling(ctx context.Context, request *Request[SamplingRequest]) (*Response[SamplingResponse], error)
}

type UnimplementedClient struct{}

func (u *UnimplementedClient) Sampling(ctx context.Context, request *Request[SamplingRequest]) (*Response[SamplingResponse], error) {
	return nil, fmt.Errorf("not implemented")
}

type Client struct {
	stream       Stream
	handler      ClientHandler
	router       *router
	interceptors []Interceptor

	callable *callable
}

func NewClient(stream Stream, handler ClientHandler, opts ...Option) *Client {
	c := &Client{
		stream:  stream,
		handler: handler,
		router:  newRouter(),
	}
	for _, opt := range opts {
		opt.applyToClient(c)
	}
	c.callable = &callable{
		Router:       c.router,
		Interceptors: c.interceptors,
		Stream:       c.stream,
	}
	return c
}

// sync.Once?
func (c *Client) Listen(ctx context.Context) error {
	for {
		msg, err := c.stream.Recv()
		if err != nil {
			return err
		}
		if msg.Method != nil {
			// c.handler.Handle(ctx, msg)
		} else {
			id, err := strconv.ParseUint(msg.ID.String(), 10, 64)
			if err != nil {
				continue
			}
			if inbox, ok := c.router.Remove(id); ok {
				inbox <- msg
			}
		}
	}
}

func (c *Client) Initialize(ctx context.Context, request *Request[InitializeRequest]) (*Response[InitializeResponse], error) {
	return call[InitializeRequest, InitializeResponse](ctx, c.callable, "initialize", request)
}

func (c *Client) ListResources(ctx context.Context, request *Request[ListResourcesRequest]) (*Response[ListResourcesResponse], error) {
	return call[ListResourcesRequest, ListResourcesResponse](ctx, c.callable, "resources/list", request)
}

func (c *Client) ListTools(ctx context.Context, request *Request[ListToolsRequest]) (*Response[ListToolsResponse], error) {
	return call[ListToolsRequest, ListToolsResponse](ctx, c.callable, "tools/list", request)
}

func (c *Client) CallTool(ctx context.Context, request *Request[CallToolRequest]) (*Response[CallToolResponse], error) {
	return call[CallToolRequest, CallToolResponse](ctx, c.callable, "tools/call", request)
}

func (c *Client) ListPrompts(ctx context.Context, request *Request[ListPromptsRequest]) (*Response[ListPromptsResponse], error) {
	return call[ListPromptsRequest, ListPromptsResponse](ctx, c.callable, "prompts/list", request)
}

func (c *Client) GetPrompt(ctx context.Context, request *Request[GetPromptRequest]) (*Response[GetPromptResponse], error) {
	return call[GetPromptRequest, GetPromptResponse](ctx, c.callable, "prompts/get", request)
}

func (c *Client) ReadResource(ctx context.Context, request *Request[ReadResourceRequest]) (*Response[ReadResourceResponse], error) {
	return call[ReadResourceRequest, ReadResourceResponse](ctx, c.callable, "resources/read", request)
}

func (c *Client) ListResourceTemplates(ctx context.Context, request *Request[ListResourceTemplatesRequest]) (*Response[ListResourceTemplatesResponse], error) {
	return call[ListResourceTemplatesRequest, ListResourceTemplatesResponse](ctx, c.callable, "resources/templates/list", request)
}

func (c *Client) Completion(ctx context.Context, request *Request[CompletionRequest]) (*Response[CompletionResponse], error) {
	return call[CompletionRequest, CompletionResponse](ctx, c.callable, "completion", request)
}

func (c *Client) Ping(ctx context.Context, request *Request[PingRequest]) (*Response[PingResponse], error) {
	return call[PingRequest, PingResponse](ctx, c.callable, "ping", request)
}

func (c *Client) SetLogLevel(ctx context.Context, request *Request[SetLogLevelRequest]) (*Response[SetLogLevelResponse], error) {
	return call[SetLogLevelRequest, SetLogLevelResponse](ctx, c.callable, "logging/setLevel", request)
}
