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
