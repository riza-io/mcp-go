package endtoend

import (
	"context"
	"io"
	"testing"

	"github.com/riza-io/mcp-go"
)

type server struct {
	mcp.UnimplementedServer
}

func (s *server) Initialize(ctx context.Context, req *mcp.Request[mcp.InitializeRequest]) (*mcp.Response[mcp.InitializeResponse], error) {
	return mcp.NewResponse(&mcp.InitializeResponse{
		ProtocolVersion: req.Params.ProtocolVersion,
	}), nil
}

func TestEndToEnd(t *testing.T) {
	ctx := context.Background()

	loggingInterceptor := mcp.UnaryInterceptorFunc(
		func(next mcp.UnaryFunc) mcp.UnaryFunc {
			return mcp.UnaryFunc(func(ctx context.Context, request mcp.AnyRequest) (mcp.AnyResponse, error) {
				t.Logf("calling: %s", request.Method())
				t.Logf("request: %s", request.ID())
				response, err := next(ctx, request)
				if err != nil {
					t.Logf("error: %v", err)
				} else {
					t.Logf("response: %s", response.Any())
				}
				return response, err
			})
		},
	)

	stdinr, stdinw := io.Pipe()
	stdoutr, stdoutw := io.Pipe()

	client := mcp.NewStdioClient(stdinr, stdoutw)
	srv := mcp.NewStdioServer(&server{}, mcp.WithInterceptors(loggingInterceptor))

	go func() {
		if err := srv.Listen(ctx, stdoutr, stdinw); err != nil {
			t.Fatalf("failed to listen: %v", err)
		}
	}()

	{
		resp, err := client.Initialize(ctx, mcp.NewRequest(&mcp.InitializeRequest{
			ProtocolVersion: "1.0.0",
		}))
		if err != nil {
			t.Fatalf("failed to initialize client: %v", err)
		}
		if resp.Result.ProtocolVersion != "1.0.0" {
			t.Fatalf("expected protocol version 1.0.0, got %s", resp.Result.ProtocolVersion)
		}
	}

	{
		_, err := client.Ping(ctx, mcp.NewRequest(&mcp.PingRequest{}))
		if err != nil {
			t.Fatalf("failed to ping server: %v", err)
		}
	}
}
