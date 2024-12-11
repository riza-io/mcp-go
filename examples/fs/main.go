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
	"github.com/riza-io/mcp-go/stdio"
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

	server := mcp.NewServer(stdio.NewStream(os.Stdin, os.Stdout), &FSServer{
		fs: os.DirFS(root),
	})

	if err := server.Listen(context.Background()); err != nil {
		log.Fatal(err)
	}
}
