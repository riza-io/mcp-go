package main

import (
	"context"
	"encoding/json"

	"github.com/riza-io/mcp-go"
)

const getForecastSchema = `{
  "type": "object",
  "properties": {
    "latitude": {
      "type": "number"
    },
    "longitude": {
      "type": "number"
    }
  },
  "required": ["latitude", "longitude"]
}`

type GetForecastArguments struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (s *WeatherServer) ListTools(ctx context.Context, req *mcp.Request[mcp.ListToolsRequest]) (*mcp.Response[mcp.ListToolsResponse], error) {
	return mcp.NewResponse(&mcp.ListToolsResponse{
		Tools: []mcp.Tool{
			{
				Name:        "get_forecast",
				Description: "Get weather forecast for a location.",
				InputSchema: json.RawMessage([]byte(getForecastSchema)),
			},
		},
	}), nil
}

func (s *WeatherServer) CallTool(ctx context.Context, req *mcp.Request[mcp.CallToolRequest]) (*mcp.Response[mcp.CallToolResponse], error) {
	var args GetForecastArguments
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return nil, err
	}

	points, err := fetchPoints(args.Latitude, args.Longitude)
	if err != nil {
		return nil, err
	}

	forecast, err := fetchForecast(points.Properties.Forecast)
	if err != nil {
		return nil, err
	}

	content := []mcp.Content{}
	for _, period := range forecast.Properties.Periods {
		content = append(content, mcp.Content{
			Type: "text",
			Text: period.DetailedForecast,
		})
	}

	return mcp.NewResponse(&mcp.CallToolResponse{
		Content: content,
	}), nil
}
