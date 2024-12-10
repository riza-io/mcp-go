package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/riza-io/mcp-go"
)

// List available weather resources.
func (s *WeatherServer) ListResources(ctx context.Context, req *mcp.Request[mcp.ListResourcesRequest]) (*mcp.Response[mcp.ListResourcesResponse], error) {
	return mcp.NewResponse(&mcp.ListResourcesResponse{
		Resources: []mcp.Resource{
			{
				URI:         "weather://" + s.defaultCity + "/current",
				Name:        "Current weather in " + s.defaultCity,
				Description: "Real-time weather data",
				MimeType:    "application/json",
			},
		},
	}), nil
}

func (s *WeatherServer) ReadResource(ctx context.Context, req *mcp.Request[mcp.ReadResourceRequest]) (*mcp.Response[mcp.ReadResourceResponse], error) {
	city := s.defaultCity

	if strings.HasPrefix(req.Params.URI, "weather://") && strings.HasSuffix(req.Params.URI, "/current") {
		city = strings.TrimPrefix(req.Params.URI, "weather://")
		city = strings.TrimSuffix(city, "/current")
	} else {
		return nil, fmt.Errorf("unknown resource: %s", req.Params.URI)
	}

	data, err := s.fetchWeather(city)
	if err != nil {
		return nil, err
	}

	contents, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return mcp.NewResponse(&mcp.ReadResourceResponse{
		Contents: []mcp.ResourceContent{
			{
				URI:      req.Params.URI,
				MimeType: "application/json",
				Text:     string(contents),
			},
		},
	}), nil
}
