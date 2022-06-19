package weather_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"weather"

	"github.com/google/go-cmp/cmp"
)

func TestParseWeatherResponse(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/weather_25.json")
	if err != nil {
		t.Fatal(err)
	}
	want := weather.Conditions{
		Summary:     "Clear",
		Temperature: 290.8,
	}
	got, err := weather.ParseWeatherResponse(data)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestParseWeatherResponseEmpty(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/weather_25_invalid.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = weather.ParseWeatherResponse(data)
	if err == nil {
		t.Fatal("want error parsing invalid response, but got nil")
	}
}

func TestParseGeoResponse(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/geo_service.json")
	if err != nil {
		t.Fatal(err)
	}
	want := weather.Coordinates{
		Lat: 50.685109,
		Lon: 7.1536622,
	}
	got, err := weather.ParseGeoResponse(data)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestParseGeoResponseEmpty(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/geo_service_invalid.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = weather.ParseGeoResponse(data)
	if err == nil {
		t.Fatal("want error parsing invalid response, but got nil")
	}
}

func TestFormatWeatherURL(t *testing.T) {
	t.Parallel()
	c := weather.NewClient("dummyAPIKey")
	location := "Paris,FR"
	want := "https://api.openweathermap.org/data/2.5/weather?q=Paris,FR&appid=dummyAPIKey"
	got := c.FormatWeatherURL(location)
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestFormatGeoURL(t *testing.T) {
	t.Parallel()
	c := weather.NewClient("dummyAPIKey")
	location := "Paris,FR"
	want := "https://api.openweathermap.org/geo/1.0/direct?q=Paris,FR&limit=1&appid=dummyAPIKey"
	got := c.FormatGeoURL(location)
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestLocationWithSpace(t *testing.T) {
	t.Parallel()
	location := []string{"HIDDEN", "What", "a", "long", "Place"}
	want := "What+a+long+Place"
	got := weather.GetLocation(location)
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

// just to check some possibilities for later tests
func TestSimpleHTTPS(t *testing.T) {
	t.Parallel()
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello, client")
		}))
	defer ts.Close()
	client := ts.Client()
	resp, err := client.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	want := http.StatusOK
	got := resp.StatusCode
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestGetWeather(t *testing.T) {
	t.Parallel()
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			f, err := os.Open("testdata/weather_25.json")
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			io.Copy(w, f)
		}))
	defer ts.Close()
	c := weather.NewClient("dummyAPIKey")
	c.BaseURL = ts.URL
	c.HTTPClient = ts.Client()
	want := weather.Conditions{
		Summary:     "Clear",
		Temperature: 290.8,
	}
	got, err := c.GetWeather("Paris,FR")
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestCelsius(t *testing.T) {
	t.Parallel()
	input := weather.Temperature(274.15)
	want := 1.0
	got := input.Celsius()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestGetCoordinates(t *testing.T) {
	t.Parallel()
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			f, err := os.Open("testdata/geo_service.json")
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			io.Copy(w, f)
		}))
	defer ts.Close()
	c := weather.NewClient("dummyAPIKey")
	c.BaseURL = ts.URL
	c.HTTPClient = ts.Client()
	want := weather.Coordinates{
		Lat: 50.685109,
		Lon: 7.1536622,
	}
	got, err := c.GetCoordinates("Paris,FR")
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
