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
	"flag"
	"io/fs"
	"log"
	"mime"
	"os"
	"path/filepath"

	"github.com/riza-io/mcp-go"
)

type FSServer struct {
	fs fs.FS

	mcp.UnimplementedServer
}

func (s *FSServer) Initialize(ctx context.Context, req *mcp.Request[mcp.InitializeRequest]) (*mcp.Response[mcp.InitializeResponse], error) {
	return mcp.NewResponse(&mcp.InitializeResponse{
		ProtocolVersion: req.Params.ProtocolVersion,
		Capabilities: mcp.ServerCapabilities{
			Resources: &mcp.Resources{},
		},
	}), nil
}

func (s *FSServer) ListResources(ctx context.Context, req *mcp.Request[mcp.ListResourcesRequest]) (*mcp.Response[mcp.ListResourcesResponse], error) {
	var resources []mcp.Resource
	fs.WalkDir(s.fs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		resources = append(resources, mcp.Resource{
			URI:      "file://" + path,
			Name:     d.Name(),
			MimeType: mime.TypeByExtension(filepath.Ext(path)),
		})
		return nil
	})
	return mcp.NewResponse(&mcp.ListResourcesResponse{
		Resources: resources,
	}), nil
}

func (s *FSServer) ReadResource(ctx context.Context, req *mcp.Request[mcp.ReadResourceRequest]) (*mcp.Response[mcp.ReadResourceResponse], error) {
	contents, err := fs.ReadFile(s.fs, req.Params.URI)
	if err != nil {
		return nil, err
	}
	return mcp.NewResponse(&mcp.ReadResourceResponse{
		Contents: []mcp.ResourceContent{
			{
				URI:      req.Params.URI,
				MimeType: mime.TypeByExtension(filepath.Ext(req.Params.URI)),
				Text:     string(contents), // TODO: base64 encode
			},
		},
	}), nil
}

func main() {
	root := flag.String("root", "/", "root directory")
	flag.Parse()

	server := mcp.NewStdioServer(&FSServer{
		fs: os.DirFS(*root),
	})

	if err := server.Listen(context.Background(), os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
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
