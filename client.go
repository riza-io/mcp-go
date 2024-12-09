package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"

	"github.com/riza-io/mcp-go/internal/jsonrpc"
)

type Client interface {
	Initialize(ctx context.Context, request *Request[InitializeRequest]) (*Response[InitializeResponse], error)
	ListResources(ctx context.Context, request *Request[ListResourcesRequest]) (*Response[ListResourcesResponse], error)
	ListTools(ctx context.Context, req *Request[ListToolsRequest]) (*Response[ListToolsResponse], error)
	CallTool(ctx context.Context, req *Request[CallToolRequest]) (*Response[CallToolResponse], error)
	ListPrompts(ctx context.Context, req *Request[ListPromptsRequest]) (*Response[ListPromptsResponse], error)
	GetPrompt(ctx context.Context, req *Request[GetPromptRequest]) (*Response[GetPromptResponse], error)
	ReadResource(ctx context.Context, req *Request[ReadResourceRequest]) (*Response[ReadResourceResponse], error)
	ListResourceTemplates(ctx context.Context, req *Request[ListResourceTemplatesRequest]) (*Response[ListResourceTemplatesResponse], error)
	Completion(ctx context.Context, req *Request[CompletionRequest]) (*Response[CompletionResponse], error)
}

type StdioClient struct {
	in           io.Reader
	out          io.Writer
	scanner      *bufio.Scanner
	next         int
	lock         sync.Mutex
	interceptors []Interceptor
}

func NewStdioClient(stdin io.Reader, stdout io.Writer, opts ...Option) Client {
	c := &StdioClient{
		in:      stdin,
		out:     stdout,
		scanner: bufio.NewScanner(stdin),
	}

	for _, opt := range opts {
		opt.applyToClient(c)
	}

	return c
}

func callUnary[P any, R any](ctx context.Context, c *StdioClient, method string, req *Request[P]) (*Response[R], error) {
	// Ensure that we are not sending multiple requests at the same time
	c.lock.Lock()
	defer c.lock.Unlock()

	defer func() {
		// Increment the ID counter
		c.next++
	}()

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

		msg := jsonrpc.Request{
			ID:      json.Number(request.ID()),
			JsonRPC: "2.0",
			Method:  request.Method(),
			Params:  json.RawMessage(rawmsg),
		}

		bs, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}

		fmt.Fprintln(c.out, string(bs))

		var result R

		for c.scanner.Scan() {
			line := c.scanner.Bytes()

			var resp jsonrpc.Response

			if err := json.Unmarshal(line, &resp); err != nil {
				return nil, err
			}

			if resp.Error != nil {
				return nil, NewError(resp.Error.Code, errors.New(resp.Error.Message))
			}

			if err := json.Unmarshal(resp.Result, &result); err != nil {
				return nil, err
			}

			break
		}

		if err := c.scanner.Err(); err != nil {
			return nil, err
		}

		return NewResponse(&result), nil
	})

	req.id = strconv.Itoa(c.next)
	req.method = method

	resp, err := interceptor.WrapUnary(inner)(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.(*Response[R]), nil
}

func (c *StdioClient) Initialize(ctx context.Context, request *Request[InitializeRequest]) (*Response[InitializeResponse], error) {
	return callUnary[InitializeRequest, InitializeResponse](ctx, c, "initialize", request)
}

func (c *StdioClient) ListResources(ctx context.Context, request *Request[ListResourcesRequest]) (*Response[ListResourcesResponse], error) {
	return callUnary[ListResourcesRequest, ListResourcesResponse](ctx, c, "resources/list", request)
}

func (c *StdioClient) ListTools(ctx context.Context, request *Request[ListToolsRequest]) (*Response[ListToolsResponse], error) {
	return callUnary[ListToolsRequest, ListToolsResponse](ctx, c, "tools/list", request)
}

func (c *StdioClient) CallTool(ctx context.Context, request *Request[CallToolRequest]) (*Response[CallToolResponse], error) {
	return callUnary[CallToolRequest, CallToolResponse](ctx, c, "tools/call", request)
}

func (c *StdioClient) ListPrompts(ctx context.Context, request *Request[ListPromptsRequest]) (*Response[ListPromptsResponse], error) {
	return callUnary[ListPromptsRequest, ListPromptsResponse](ctx, c, "prompts/list", request)
}

func (c *StdioClient) GetPrompt(ctx context.Context, request *Request[GetPromptRequest]) (*Response[GetPromptResponse], error) {
	return callUnary[GetPromptRequest, GetPromptResponse](ctx, c, "prompts/get", request)
}

func (c *StdioClient) ReadResource(ctx context.Context, request *Request[ReadResourceRequest]) (*Response[ReadResourceResponse], error) {
	return callUnary[ReadResourceRequest, ReadResourceResponse](ctx, c, "resources/read", request)
}

func (c *StdioClient) ListResourceTemplates(ctx context.Context, request *Request[ListResourceTemplatesRequest]) (*Response[ListResourceTemplatesResponse], error) {
	return callUnary[ListResourceTemplatesRequest, ListResourceTemplatesResponse](ctx, c, "resources/templates/list", request)
}

func (c *StdioClient) Completion(ctx context.Context, request *Request[CompletionRequest]) (*Response[CompletionResponse], error) {
	return callUnary[CompletionRequest, CompletionResponse](ctx, c, "completion", request)
}
