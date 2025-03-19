package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func fetch(url string) (*json.RawMessage, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "mcp-go-weather/1.0")
	req.Header.Set("Accept", "application/geo+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	var data json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

type Points struct {
	Properties struct {
		Forecast string `json:"forecast"`
	} `json:"properties"`
}

func fetchPoints(lat, lon float64) (*Points, error) {
	url := fmt.Sprintf("https://api.weather.gov/points/%f,%f", lat, lon)
	data, err := fetch(url)
	if err != nil {
		return nil, err
	}
	var points Points
	if err := json.Unmarshal(*data, &points); err != nil {
		return nil, err
	}
	return &points, nil
}

type Forecast struct {
	Properties struct {
		Periods []Period `json:"periods"`
	} `json:"properties"`
}

type Period struct {
	Name             string  `json:"name"`
	Temp             float64 `json:"temperature"`
	TempUnit         string  `json:"temperatureUnit"`
	WindSpeed        string  `json:"windSpeed"`
	WindDirection    string  `json:"windDirection"`
	DetailedForecast string  `json:"detailedForecast"`
}

func fetchForecast(url string) (*Forecast, error) {
	data, err := fetch(url)
	if err != nil {
		return nil, err
	}
	fmt.Println("forecast", string(*data))
	var forecast Forecast
	if err := json.Unmarshal(*data, &forecast); err != nil {
		return nil, err
	}
	return &forecast, nil
}
