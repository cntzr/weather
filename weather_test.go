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

func TestConditionsFromParseWeatherResponse(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/weather_30.json")
	if err != nil {
		t.Fatal(err)
	}
	want := weather.Conditions{
		Summary:       "Leichter Regen",
		Temperature:   31.38,
		Timestamp:     "17.06.2022 17:23 CEST",
		Sunrise:       "05:18",
		Sunset:        "21:46",
		FeelsLike:     29.86,
		DewPoint:      10.15,
		Pressure:      1021,
		Humidity:      27,
		WindSpeed:     2.3,
		WindGust:      3.32,
		WindDirection: 233,
	}
	got, _, err := weather.ParseWeatherResponse(data)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestDailyForecastFromParseWeatherResponse(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/weather_30.json")
	if err != nil {
		t.Fatal(err)
	}
	want := weather.ForecastDaily{
		Day:       "17.06.2022",
		Moonrise:  "00:24",
		Moonset:   "08:14",
		Moonphase: 0.62,
		Temp: weather.DailyTempBenchmarks{
			Max:     31.38,
			Min:     13.58,
			Morning: 15.53,
			Day:     28.02,
			Evening: 30.18,
			Night:   20.39,
		},
		Alerts: []weather.Alert{},
	}
	_, fc, err := weather.ParseWeatherResponse(data)
	if err != nil {
		t.Fatal(err)
	}
	// we are looking only for the first set of moon data to improve the current conditions
	got := fc.Daily[0]
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestForecastFromParseWeatherResponseEmpty(t *testing.T) {
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

func TestConditionsFromParseWeatherResponseEmpty(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/weather_30_invalid.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = weather.ParseWeatherResponse(data)
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
		Lat: 55.123456,
		Lon: 3.7654321,
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
	coordinates := weather.Coordinates{
		Lat: 55.123456,
		Lon: 3.7654321,
	}
	want := "https://api.openweathermap.org/data/3.0/onecall?lat=55.123456&lon=3.7654321&units=metric&lang=de&appid=dummyAPIKey"
	got := c.FormatWeatherURL(coordinates)
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
	params := []string{"HIDDEN", "HIDDEN", "What", "a", "long", "Place"}
	want := "What+a+long+Place"
	got := weather.GetLocation(params)
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestFunctionalParameter(t *testing.T) {
	t.Parallel()
	params := []string{"HIDDEN", "doit", "HIDDEN", "HIDDEN"}
	want := "doit"
	got := weather.GetFunction(params)
	if want != got {
		t.Errorf("want %s, got %s", want, got)
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

func TestConditionsFromGetWeather(t *testing.T) {
	t.Parallel()
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			f, err := os.Open("testdata/weather_30.json")
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
		Summary:       "Leichter Regen",
		Temperature:   31.38,
		Timestamp:     "17.06.2022 17:23 CEST",
		Sunrise:       "05:18",
		Sunset:        "21:46",
		FeelsLike:     29.86,
		DewPoint:      10.15,
		Pressure:      1021,
		Humidity:      27,
		WindSpeed:     2.3,
		WindGust:      3.32,
		WindDirection: 233,
	}
	coordinates := weather.Coordinates{Lat: 1.0, Lon: 2.0}
	got, _, err := c.GetWeather(coordinates)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestForecastHourlyFromGetWeather(t *testing.T) {
	t.Parallel()
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			f, err := os.Open("testdata/weather_30.json")
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
	want := weather.ForecastHourly{
		Day:         "17.06.2022",
		Hour:        "17:00",
		Temperature: 31.38,
	}
	coordinates := weather.Coordinates{Lat: 1.0, Lon: 2.0}
	_, fc, err := c.GetWeather(coordinates)
	if err != nil {
		t.Fatal(err)
	}
	got := fc.Hourly[0]
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestForecastDailyFromGetWeather(t *testing.T) {
	t.Parallel()
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			f, err := os.Open("testdata/weather_30.json")
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
	want := weather.ForecastDaily{
		Day:       "17.06.2022",
		Moonrise:  "00:24",
		Moonset:   "08:14",
		Moonphase: 0.62,
		Temp: weather.DailyTempBenchmarks{
			Max:     31.38,
			Min:     13.58,
			Morning: 15.53,
			Day:     28.02,
			Evening: 30.18,
			Night:   20.39,
		},
		Alerts: []weather.Alert{},
	}
	coordinates := weather.Coordinates{Lat: 1.0, Lon: 2.0}
	_, fc, err := c.GetWeather(coordinates)
	if err != nil {
		t.Fatal(err)
	}
	// we are looking only for the first set of moon data to improve the current conditions
	got := fc.Daily[0]
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestPrintForcastWithWrongOffset(t *testing.T) {
	t.Parallel()
	err := weather.PrintForecast(weather.Forecast{}, 9)
	if err == nil {
		t.Errorf("want error for wrong offset, but got nil")
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
		Lat: 55.123456,
		Lon: 3.7654321,
	}
	got, err := c.GetCoordinates("Paris,FR")
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestGetTimestamp(t *testing.T) {
	t.Parallel()
	// TODO Testserie mit verschiedenen Ausgaben aufbauen
	want := "17.06.2022 17:23 CEST"
	got := weather.GetTimestamp(1655479384, "02.01.2006 15:04 MST")
	if want != got {
		t.Errorf("want %s, got %s", want, got)
	}
}

func TestSpeedInKilometres(t *testing.T) {
	t.Parallel()
	input := weather.Speed(10.0)
	want := 36.0
	got := input.KmPerHour()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestDirection(t *testing.T) {
	t.Parallel()
	input := weather.Direction(190.0)
	want := "S"
	got := input.Direction()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
