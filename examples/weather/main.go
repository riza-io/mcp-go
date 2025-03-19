package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/riza-io/mcp-go"
	"github.com/riza-io/mcp-go/sse"
)

type WeatherServer struct {
	mcp.UnimplementedServer
}

func (s *WeatherServer) Initialize(ctx context.Context, req *mcp.Request[mcp.InitializeRequest]) (*mcp.Response[mcp.InitializeResponse], error) {
	fmt.Println("Initialize", req.Params.ProtocolVersion)
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

	mux := http.NewServeMux()

	server := mcp.NewServer(sse.NewStream(mux, "/sse", "/messages"), &WeatherServer{})

	go func() {
		if err := http.ListenAndServe(":3009", mux); err != nil {
			log.Fatal(err)
		}
	}()

	if err := server.Listen(ctx); err != nil {
		log.Fatal(err)
	}
}
