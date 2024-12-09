package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/riza-io/mcp-go/internal/jsonrpc"
)

type Server interface {
	Initialize(ctx context.Context, req *Request[InitializeRequest]) (*Response[InitializeResponse], error)
	ListTools(ctx context.Context, req *Request[ListToolsRequest]) (*Response[ListToolsResponse], error)
	CallTool(ctx context.Context, req *Request[CallToolRequest]) (*Response[CallToolResponse], error)
	ListPrompts(ctx context.Context, req *Request[ListPromptsRequest]) (*Response[ListPromptsResponse], error)
	GetPrompt(ctx context.Context, req *Request[GetPromptRequest]) (*Response[GetPromptResponse], error)
	ListResources(ctx context.Context, req *Request[ListResourcesRequest]) (*Response[ListResourcesResponse], error)
	ReadResource(ctx context.Context, req *Request[ReadResourceRequest]) (*Response[ReadResourceResponse], error)
	ListResourceTemplates(ctx context.Context, req *Request[ListResourceTemplatesRequest]) (*Response[ListResourceTemplatesResponse], error)
	Completion(ctx context.Context, req *Request[CompletionRequest]) (*Response[CompletionResponse], error)
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

func process[T, V any](ctx context.Context, cfg *serverConfig, msg jsonrpc.Request, params *T, method func(ctx context.Context, req *Request[T]) (*Response[V], error)) (any, error) {
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

	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}
	req := NewRequest(params)
	req.id = msg.ID.String()
	req.method = msg.Method

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
		return nil, err
	}

	resp := rr.(*Response[V])
	return resp.Result, nil

}

type serverConfig struct {
	interceptors []Interceptor
}

type StdioServer struct {
	cfg *serverConfig
	srv Server
}

func NewStdioServer(srv Server, opts ...Option) *StdioServer {
	cfg := &serverConfig{}
	for _, opt := range opts {
		opt.applyToServer(cfg)
	}
	return &StdioServer{
		cfg: cfg,
		srv: srv,
	}
}

func (s StdioServer) Listen(ctx context.Context, r io.Reader, w io.Writer) error {
	cfg := s.cfg
	srv := s.srv

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		dec := json.NewDecoder(strings.NewReader(line))

		var msg jsonrpc.Request
		if err := dec.Decode(&msg); err != nil {
			continue
		}

		var result any
		var err error
		code := 9

		switch msg.Method {
		case "initialize":
			params := &InitializeRequest{}
			result, err = process(ctx, cfg, msg, params, srv.Initialize)
		case "completion/complete":
			params := &CompletionRequest{}
			result, err = process(ctx, cfg, msg, params, srv.Completion)
		case "tools/list":
			params := &ListToolsRequest{}
			result, err = process(ctx, cfg, msg, params, srv.ListTools)
		case "tools/call":
			params := &CallToolRequest{}
			result, err = process(ctx, cfg, msg, params, srv.CallTool)
		case "prompts/list":
			params := &ListPromptsRequest{}
			result, err = process(ctx, cfg, msg, params, srv.ListPrompts)
		case "prompts/get":
			params := &GetPromptRequest{}
			result, err = process(ctx, cfg, msg, params, srv.GetPrompt)
		case "resources/list":
			params := &ListResourcesRequest{}
			result, err = process(ctx, cfg, msg, params, srv.ListResources)
		case "resources/read":
			params := &ReadResourceRequest{}
			result, err = process(ctx, cfg, msg, params, srv.ReadResource)
		case "resources/templates/list":
			params := &ListResourceTemplatesRequest{}
			result, err = process(ctx, cfg, msg, params, srv.ListResourceTemplates)
		default:
			if msg.ID == "" {
				// Ignore notifications
				continue
			}
			code = -32601
			err = fmt.Errorf("unknown method: %s", msg.Method)
		}

		var resp jsonrpc.Response
		if err != nil {
			resp = jsonrpc.Response{
				ID:      msg.ID,
				JsonRPC: msg.JsonRPC,
				Error: &jsonrpc.ErrorDetail{
					Code:    code,
					Message: err.Error(),
				},
			}
		} else {
			rawresult, err := json.Marshal(result)
			if err != nil {
				return err
			}
			resp = jsonrpc.Response{
				ID:      msg.ID,
				JsonRPC: msg.JsonRPC,
				Result:  rawresult,
			}
		}

		bs, err := json.Marshal(resp)
		if err != nil {
			return err
		}

		fmt.Fprintln(w, string(bs))
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
