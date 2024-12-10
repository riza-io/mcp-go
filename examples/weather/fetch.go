package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type WeatherData struct {
	Temperature float64   `json:"temperature"`
	Conditions  string    `json:"conditions"`
	Humidity    float64   `json:"humidity"`
	WindSpeed   float64   `json:"wind_speed"`
	Timestamp   time.Time `json:"timestamp"`
}

func (s *WeatherServer) fetchWeather(city string) (*WeatherData, error) {
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

	var data WeatherData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

type Forecast struct {
	Date        string  `json:"date"`
	Temperature float64 `json:"temperature"`
	Conditions  string  `json:"conditions"`
}

func (s *WeatherServer) fetchForecast(city string, days int) ([]Forecast, error) {
	q := url.Values{}
	q.Set("q", city)
	q.Set("appid", s.key)
	q.Set("units", "metric")
	q.Set("cnt", strconv.Itoa(days))

	uri := "http://api.openweathermap.org/data/2.5/weather?" + q.Encode()
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data WeatherData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return nil, nil
}
