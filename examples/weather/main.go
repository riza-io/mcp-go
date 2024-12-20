package main

import (
	"context"
	"log"
	"os"

	"github.com/riza-io/mcp-go"
	"github.com/riza-io/mcp-go/stdio"
)

type WeatherServer struct {
	key         string
	defaultCity string

	mcp.UnimplementedServer
}

func (s *WeatherServer) Initialize(ctx context.Context, req *mcp.Request[mcp.InitializeRequest]) (*mcp.Response[mcp.InitializeResponse], error) {
	return mcp.NewResponse(&mcp.InitializeResponse{
		ProtocolVersion: req.Params.ProtocolVersion,
		Capabilities: mcp.ServerCapabilities{
			Resources: &mcp.Resources{},
			Tools:     &mcp.Tools{},
		},
	}), nil
}

func main() {
	ctx := context.Background()

	if os.Getenv("OPENWEATHER_API_KEY") == "" {
		log.Fatal("OPENWEATHER_API_KEY environment variable required")
	}

	server := mcp.NewServer(stdio.NewStream(os.Stdin, os.Stdout), &WeatherServer{
		defaultCity: "London",
		key:         os.Getenv("OPENWEATHER_API_KEY"),
	})

	if err := server.Listen(ctx); err != nil {
		log.Fatal(err)
	}
}
