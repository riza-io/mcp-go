MCP Go SDK
==========

[![Build](https://github.com/riza-io/mcp-go/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/riza-io/mcp-go/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/riza-io/mcp-go?cache)](https://goreportcard.com/report/github.com/riza-io/mcp-go)
[![GoDoc](https://pkg.go.dev/badge/github.com/riza-io/mcp-go.svg)](https://pkg.go.dev/github.com/riza-io/mcp-go)


Go implementation of the [Model Context Protocol](https://modelcontextprotocol.io) (MCP), providing both client and server capabilities for integrating with LLM surfaces.

## Overview

The Model Context Protocol allows applications to provide context for LLMs in a standardized way, separating the concerns of providing context from the actual LLM interaction. This Go SDK implements the full MCP specification, making it easy to:

- Build MCP clients that can connect to any MCP server
- Create MCP servers that expose resources, prompts and tools
- Use standard transports like stdio and SSE (coming soon)
- Handle all MCP protocol messages and lifecycle events

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
	"strings"

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
			Name:     path,
			MimeType: mime.TypeByExtension(filepath.Ext(path)),
		})
		return nil
	})
	return mcp.NewResponse(&mcp.ListResourcesResponse{
		Resources: resources,
	}), nil
}

func (s *FSServer) ReadResource(ctx context.Context, req *mcp.Request[mcp.ReadResourceRequest]) (*mcp.Response[mcp.ReadResourceResponse], error) {
	contents, err := fs.ReadFile(s.fs, strings.TrimPrefix(req.Params.URI, "file://"))
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
	flag.Parse()

	root := flag.Arg(0)
	if root == "" {
		root = "/"
	}

	server := mcp.NewStdioServer(&FSServer{
		fs: os.DirFS(root),
	})

	if err := server.Listen(context.Background(), os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
```

This example can be compiled and wired up to Claude Desktop (or any other MCP client).

```json
{
	"mcpServers": {
		"fs": {
			"command": "/path/to/mcp-go-fs",
			"args": [
				"/path/to/root/directory"
			]
		}
	}
}
```

## Documentation

- [Model Context Protocol documentation](https://modelcontextprotocol.io)
- [MCP Specification](https://spec.modelcontextprotocol.io)
- [Example Servers](https://github.com/riza-io/mcp-go/tree/main/examples)

## Roadmap

The majority of the base protocol has been implemented. The following features are on our roadmap:

- Notifications
- Sampling
- Roots

## Legal

Offered under the [MIT license][license].
