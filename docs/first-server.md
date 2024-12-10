# Create a simple MCP server in Go in 15 minutes

Let's build your first MCP server in Go! We'll create a weather server that provides current weather data as a resource and lets Claude fetch forecasts using tools.

> This guide uses the OpenWeatherMap API. You'll need a free API key from [OpenWeatherMap](https://openweathermap.org/api) to follow along.

## Prerequisites

You'll need Go 1.22 or higher

```bash
go version  # Should be 1.22 or higher
```

Create a new module using `go mod init`.

```bash
mkdir mcp-go-weather
cd mcp-go-weather
go mod init github.com/example/mcp-go-weather
```

Add your API key to the environment.

```bash
export OPENWEATHER_API_KEY=your-api-key-here
```

## Create your server

### Add the base imports and setup

In `main.go` add the entry point and base server implemenation.

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/riza-io/mcp-go"
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

	server := mcp.NewStdioServer(&WeatherServer{
		defaultCity: "London",
		key:         os.Getenv("OPENWEATHER_API_KEY"),
	})

	if err := server.Listen(ctx, os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
```

### Add weather fetching functionality

In `fetch.go` add two functions to fetch the weather and the five-day forecast.
Note that JSON handling in Go is verbose, hence the length of this code.

```go
package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Weather struct {
	Temperature float64   `json:"temperature"`
	Conditions  string    `json:"conditions"`
	Humidity    float64   `json:"humidity"`
	WindSpeed   float64   `json:"wind_speed"`
	Timestamp   time.Time `json:"timestamp"`
}

type weatherResponse struct {
	Main struct {
		Temp     float64 `json:"temp"`
		Humidity float64 `json:"humidity"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
}

func (s *WeatherServer) fetchWeather(city string) (*Weather, error) {
	q := url.Values{}
	q.Set("q", city)
	q.Set("appid", s.key)
	q.Set("units", "metric")

	uri := "http://api.openweathermap.org/data/2.5/weather?" + q.Encode()
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data weatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &Weather{
		Temperature: data.Main.Temp,
		Conditions:  data.Weather[0].Description,
		Humidity:    data.Main.Humidity,
		WindSpeed:   data.Wind.Speed,
		Timestamp:   time.Now(),
	}, nil
}

type Forecast struct {
	Date        string  `json:"date"`
	Temperature float64 `json:"temperature"`
	Conditions  string  `json:"conditions"`
}

type forecastResponse struct {
	List []struct {
		DatetimeText string `json:"dt_txt"`
		Main         struct {
			Temp float64 `json:"temp"`
		} `json:"main"`
		Weather []struct {
			Description string `json:"description"`
		} `json:"weather"`
	} `json:"list"`
}

func (s *WeatherServer) fetchForecast(city string, days int) ([]Forecast, error) {
	q := url.Values{}
	q.Set("q", city)
	q.Set("cnt", strconv.Itoa(days*8))
	q.Set("appid", s.key)
	q.Set("units", "metric")

	uri := "http://api.openweathermap.org/data/2.5/forecast?" + q.Encode()
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data forecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	forecasts := []Forecast{}
	for i, day_data := range data.List {
		if i%8 != 0 {
			continue
		}

		date, _, _ := strings.Cut(day_data.DatetimeText, " ")

		forecasts = append(forecasts, Forecast{
			Date:        date,
			Temperature: day_data.Main.Temp,
			Conditions:  day_data.Weather[0].Description,
		})
	}

	return forecasts, nil
}
```

### Implement resource handlers

Add resource-related handlers to a new `reosurces.go` file.

```go
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
```

### Implement tool handlers

Add these tool-related handlers to `tools.go`.

```go
package main

import (
	"context"
	"encoding/json"

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

	text, err := json.MarshalIndent(forecasts, "", "  ")
	if err != nil {
		return nil, err
	}

	return mcp.NewResponse(&mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: string(text),
			},
		},
	}), nil
}
```

The server is now complete! Build it using `go build`

```bash
go build -o mcp-weather ./...
```

## Connect to Claude Desktop

### Update Claude config

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "weather": {
      "command": "/path/to/mcp-weather",
      "env": {
        "OPENWEATHER_API_KEY": "your-api-key"
      }
    }
  }
}
```

### Restart Claude

1. Quit Claude completely
2. Start Claude again
3. Look for your weather server in the 🔌 menu

## Try it out!

Ask Claude the following questions:

> What's the current weather in San Francisco? Can you analyze the conditions and tell me if it's a good day for outdoor activities?


> Can you get me a 5-day forecast for Tokyo and help me plan what clothes to pack for my trip?

> Can you analyze the forecast for both Tokyo and San Francisco and tell me which city would be better for outdoor photography this week?

## Available transports

mcp-go currently supports the stdio transport. Follow this
[issue](https://github.com/riza-io/mcp-go/issues/5) to track progress on the SSE
transports.

## Next steps

- [Architecture overview](https://docs.riza.io/concepts/architecture)
- [MCP Go SDK](https://github.com/riza-io/mcp-go)