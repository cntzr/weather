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

	OWMResponse struct {
		Weather []struct {
			Main string
		}
		Main struct {
			Temp Temperature
		}
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

func ParseResponse(data []byte) (Conditions, error) {
	var resp OWMResponse
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return Conditions{}, fmt.Errorf("invalid API response %s: %w", data, err)
	}
	if len(resp.Weather) < 1 {
		return Conditions{}, fmt.Errorf("invalid API response %s: want at lest one Weather element", data)
	}
	conditions := Conditions{
		Summary:     resp.Weather[0].Main,
		Temperature: resp.Main.Temp,
	}
	return conditions, nil
}

func (t Temperature) Celsius() float64 {
	return float64(t) - 273.15
}

func (c *Client) FormatURL(location string) string {
	return fmt.Sprintf("%s/data/2.5/weather?q=%s&appid=%s", c.BaseURL, location, c.APIKey)
}

func (c *Client) GetWeather(location string) (Conditions, error) {
	URL := c.FormatURL(location)
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
	conditions, err := ParseResponse(data)
	if err != nil {
		return Conditions{}, err
	}
	return conditions, nil
}
