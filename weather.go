package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type (
	Client struct {
		APIKey     string
		BaseURL    string
		HTTPClient *http.Client
	}

	Conditions struct {
		Summary     string
		Temperature Temperature
	}

	Coordinates struct {
		Lon float64
		Lat float64
	}

	WeatherResponse struct {
		Weather []struct {
			Main string
		}
		Main struct {
			Temp Temperature
		}
	}

	GeoResponse []struct {
		Lon float64
		Lat float64
	}

	Temperature float64
)

func RunCLI() {
	key := os.Getenv("OPENWEATHERMAP_API_KEY")
	if key == "" {
		fmt.Fprintln(os.Stderr, "Please set the env variable OPENWEATHERMAP_API_KEY")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s LOCATION\n\nExample: %[1]s London,UK\n", os.Args[0])
		os.Exit(1)
	}

	location := GetLocation(os.Args)
	c := NewClient(key)
	conditions, err := c.GetWeather(location)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%s - %.1f °C\n", conditions.Summary, conditions.Temperature.Celsius())
}

func GetLocation(args []string) string {
	return strings.Join(args[1:], "+")
}

func Get(location, key string) (Conditions, error) {
	c := NewClient(key)
	conditions, err := c.GetWeather(location)
	if err != nil {
		return Conditions{}, err
	}
	return conditions, nil
}

func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:  apiKey,
		BaseURL: "https://api.openweathermap.org",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func ParseWeatherResponse(data []byte) (Conditions, error) {
	var resp WeatherResponse
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return Conditions{}, fmt.Errorf("invalid API response %s: %w", data, err)
	}
	if len(resp.Weather) < 1 {
		return Conditions{}, fmt.Errorf("invalid API response %s: want at least one Weather element", data)
	}
	conditions := Conditions{
		Summary:     resp.Weather[0].Main,
		Temperature: resp.Main.Temp,
	}
	return conditions, nil
}

func ParseGeoResponse(data []byte) (Coordinates, error) {
	var resp GeoResponse
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return Coordinates{}, fmt.Errorf("invalid API response %s: %w", data, err)
	}
	if len(resp) < 1 {
		return Coordinates{}, fmt.Errorf("invalid API response %s: want at least one set of coordinates", data)
	}
	coordinates := Coordinates{
		Lat: resp[0].Lat,
		Lon: resp[0].Lon,
	}
	return coordinates, nil
}

func (t Temperature) Celsius() float64 {
	return float64(t) - 273.15
}

func (c *Client) FormatWeatherURL(location string) string {
	// TODO ... beim Refactoring unit=metric und lang=de einbauen
	return fmt.Sprintf("%s/data/2.5/weather?q=%s&appid=%s", c.BaseURL, location, c.APIKey)
}

func (c *Client) FormatGeoURL(location string) string {
	return fmt.Sprintf("%s/geo/1.0/direct?q=%s&limit=1&appid=%s", c.BaseURL, location, c.APIKey)
}

func (c *Client) GetWeather(location string) (Conditions, error) {
	URL := c.FormatWeatherURL(location)
	resp, err := c.HTTPClient.Get(URL)
	if err != nil {
		return Conditions{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Conditions{}, fmt.Errorf("unexptected response status %q", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Conditions{}, err
	}
	conditions, err := ParseWeatherResponse(data)
	if err != nil {
		return Conditions{}, err
	}
	return conditions, nil
}

func (c *Client) GetCoordinates(location string) (Coordinates, error) {
	URL := c.FormatGeoURL(location)
	resp, err := c.HTTPClient.Get(URL)
	if err != nil {
		return Coordinates{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Coordinates{}, fmt.Errorf("unexptected response status %q", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Coordinates{}, err
	}
	coordinates, err := ParseGeoResponse(data)
	if err != nil {
		return Coordinates{}, err
	}
	return coordinates, nil
}
