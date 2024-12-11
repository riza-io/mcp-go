package endtoend

import (
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/riza-io/mcp-go"
	"github.com/riza-io/mcp-go/stdio"
)

type server struct {
	mcp.UnimplementedServer
}

type client struct {
	mcp.UnimplementedClient
}

func (s *server) Initialize(ctx context.Context, req *mcp.Request[mcp.InitializeRequest]) (*mcp.Response[mcp.InitializeResponse], error) {
	return mcp.NewResponse(&mcp.InitializeResponse{
		ProtocolVersion: req.Params.ProtocolVersion,
	}), nil
}

func (s *server) SetLogLevel(ctx context.Context, req *mcp.Request[mcp.SetLogLevelRequest]) (*mcp.Response[mcp.SetLogLevelResponse], error) {
	return mcp.NewResponse(&mcp.SetLogLevelResponse{}), nil
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

	c := mcp.NewClient(stdio.NewStream(stdinr, stdoutw), &client{},
		mcp.WithInterceptors(loggingInterceptor))
	s := mcp.NewServer(stdio.NewStream(stdoutr, stdinw), &server{})

	go func() {
		if err := s.Listen(ctx); err != nil {
			t.Fatalf("failed to listen: %v", err)
		}
	}()

	go func() {
		if err := c.Listen(ctx); err != nil {
			t.Fatalf("failed to listen: %v", err)
		}
	}()

	t.Run("initialize", func(t *testing.T) {
		resp, err := c.Initialize(ctx, mcp.NewRequest(&mcp.InitializeRequest{
			ProtocolVersion: "1.0.0",
		}))
		if err != nil {
			t.Fatalf("failed to initialize client: %v", err)
		}
		if resp.Result.ProtocolVersion != "1.0.0" {
			t.Fatalf("expected protocol version 1.0.0, got %s", resp.Result.ProtocolVersion)
		}
	})

	t.Run("client/ping", func(t *testing.T) {
		_, err := c.Ping(ctx, mcp.NewRequest(&mcp.PingRequest{}))
		if err != nil {
			t.Fatalf("failed to ping server: %v", err)
		}
	})

	t.Run("server/ping", func(t *testing.T) {
		_, err := s.Ping(ctx, mcp.NewRequest(&mcp.PingRequest{}))
		if err != nil {
			t.Fatalf("failed to ping client: %v", err)
		}
	})

	t.Run("server/sendLogMessage", func(t *testing.T) {
		err := s.LogMessage(ctx, mcp.NewRequest(&mcp.LogMessageRequest{
			Level:  mcp.LevelInfo,
			Logger: "test",
			Data:   json.RawMessage(`{"message": "test"}`),
		}))
		if err != nil {
			t.Fatalf("failed to send log message: %v", err)
		}
	})

	t.Run("set log level", func(t *testing.T) {
		_, err := c.SetLogLevel(ctx, mcp.NewRequest(&mcp.SetLogLevelRequest{
			Level: mcp.LevelInfo,
		}))
		if err != nil {
			t.Fatalf("failed to set log level: %v", err)
		}
	})
}
