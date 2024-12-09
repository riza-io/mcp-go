mcp-go
=======

[![Build](https://github.com/riza-io/mcp-go/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/riza-io/mcp-go/actions/workflows/ci.yaml)
[![Report Card](https://goreportcard.com/badge/github.com/riza-io/mcp-go)](https://goreportcard.com/report/github.com/riza-io/mcp-go)
[![GoDoc](https://pkg.go.dev/badge/github.com/riza-io/mcp-go.svg)](https://pkg.go.dev/github.com/riza-io/mcp-go)

mcp-go is a Go implementation of the [Model Context
Protocol](https://modelcontextprotocol.io/introduction). The client and server
support resources, prompts and tools. Support for sampling and roots in on the
roadmap.

## A small example

Curious what all this looks like in practice? Here's an example server that
exposes the contents of a `io.FS` as resources.

```go
package main

import (
  "context"
  "log"
  "net/http"

  "connectrpc.com/connect"
  pingv1 "connectrpc.com/connect/internal/gen/connect/ping/v1"
  "connectrpc.com/connect/internal/gen/connect/ping/v1/pingv1connect"
  "golang.org/x/net/http2"
  "golang.org/x/net/http2/h2c"
)

type PingServer struct {
  pingv1connect.UnimplementedPingServiceHandler // returns errors from all methods
}

func (ps *PingServer) Ping(
  ctx context.Context,
  req *connect.Request[pingv1.PingRequest],
) (*connect.Response[pingv1.PingResponse], error) {
  // connect.Request and connect.Response give you direct access to headers and
  // trailers. No context-based nonsense!
  log.Println(req.Header().Get("Some-Header"))
  res := connect.NewResponse(&pingv1.PingResponse{
    // req.Msg is a strongly-typed *pingv1.PingRequest, so we can access its
    // fields without type assertions.
    Number: req.Msg.Number,
  })
  res.Header().Set("Some-Other-Header", "hello!")
  return res, nil
}

func main() {
  mux := http.NewServeMux()
  // The generated constructors return a path and a plain net/http
  // handler.
  mux.Handle(pingv1connect.NewPingServiceHandler(&PingServer{}))
  err := http.ListenAndServe(
    "localhost:8080",
    // For gRPC clients, it's convenient to support HTTP/2 without TLS. You can
    // avoid x/net/http2 by using http.ListenAndServeTLS.
    h2c.NewHandler(mux, &http2.Server{}),
  )
  log.Fatalf("listen failed: %v", err)
}
```

This example can be compiled to a static binary and wired up to Claude Desktop
or any other MCP client.

Screenshot?

## Roadmap

The majority of the base protocol has been implemented. The following features are on our roadmap:

- Notifications
- Sampling
- Roots

## Legal

Offered under the [MIT license][license].
