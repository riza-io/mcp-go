package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/riza-io/mcp-go"
)

const getForecastSchema = `{
  "type": "object",
  "properties": {
    "city": {
      "type": "string",
      "description": "City name"
    },
    "days": {
      "type": "number",
      "description": "Number of days (1-5)",
      "minimum": 1,
      "maximum": 5
    }
  },
  "required": ["city"]
}`

type GetForecastArguments struct {
	City string `json:"city"`
	Days int    `json:"days"`
}

func (s *WeatherServer) ListTools(ctx context.Context, req *mcp.Request[mcp.ListToolsRequest]) (*mcp.Response[mcp.ListToolsResponse], error) {
	return mcp.NewResponse(&mcp.ListToolsResponse{
		Tools: []mcp.Tool{
			{
				Name:        "get_forecast",
				Description: "Get weather forecast for a city",
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

	forecasts, err := s.fetchForecast(args.City, args.Days)
	if err != nil {
		return nil, err
	}

	return mcp.NewResponse(&mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("Forecast for %s for the next %d days:\n%v", args.City, args.Days, forecasts),
			},
		},
	}), nil
}
