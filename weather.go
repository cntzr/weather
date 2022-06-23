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

	Coordinates struct {
		Lon float64
		Lat float64
	}

	Conditions struct {
		Timestamp     string
		Sunrise       string
		Sunset        string
		Summary       string
		Temperature   float64
		FeelsLike     float64
		DewPoint      float64
		Pressure      int
		Humidity      int
		WindSpeed     Speed
		WindGust      Speed
		WindDirection Direction
	}

	ForecastHourly struct {
		Timestamp string
	}

	ForecastDaily struct {
		Timestamp string
		Moonrise  string
		Moonset   string
		Moonphase Phase
	}

	Forecast struct {
		Hourly []ForecastHourly
		Daily  []ForecastDaily
	}

	WeatherResponse struct {
		Current struct {
			Weather []struct {
				Description string
			}
			DT         int64
			Sunrise    int64
			Sunset     int64
			Temp       float64
			Feels_Like float64
			Dew_Point  float64
			Pressure   int
			Humidity   int
			Wind_Speed Speed
			Wind_Gust  Speed
			Wind_Deg   Direction
		}
		Hourly []struct {
			DT int64
		}
		Daily []struct {
			DT         int64
			Moonrise   int64
			Moonset    int64
			Moon_Phase Phase
		}
	}

	GeoResponse []struct {
		Lon float64
		Lat float64
	}

	Speed float64

	Direction float64

	Phase float64
)

const (
	// limits for wind directions
	N   = 0.0   // N ... Norden
	NNO = 22.5  // NNO ... NordNordOsten
	NO  = 45.0  // NO ... NordOsten
	ONO = 67.5  // ONO ... OstNordOsten
	O   = 90.0  // O ... Osten
	OSO = 112.5 // OSO ... OstSüdOsten
	SO  = 135.0 // SO ... SüdOsten
	SSO = 157.5 // SSO ... SüdSüdOsten
	S   = 180.0 // S ... Süden
	SSW = 202.5 // SSW ... SüdSüdWesten
	SW  = 225.0 // SW ... SüdWesten
	WSW = 247.5 // WSW ... WestSüdWesten
	W   = 270.0 // W ... Westen
	WNW = 292.5 // WNW ... WestNordWesten
	NW  = 315.0 // NW ... NordWesten
	NNW = 337.5 // NNW ... NordNordWesten

	// function arguments for CLI
	FunctionCurrent       = "current"
	FunctionToday         = "today"
	FunctionTomorrow      = "tomorrow"
	FunctionAfterTomorrow = "aftertomorrow"
	FunctionMoon          = "moon"
	FunctionRain          = "rain"
	FunctionAlert         = "alert"
)

var validFunction = map[string]bool{
	FunctionCurrent:       true,
	FunctionToday:         true,
	FunctionTomorrow:      true,
	FunctionAfterTomorrow: true,
	FunctionMoon:          true,
	FunctionRain:          true,
	FunctionAlert:         true,
}

func RunCLI() {
	key := os.Getenv("OPENWEATHERMAP_API_KEY")
	if key == "" {
		fmt.Fprintln(os.Stderr, "Please set the env variable OPENWEATHERMAP_API_KEY")
		os.Exit(1)
	}

	if len(os.Args) < 3 || !validFunction[os.Args[1]] {
		fmt.Fprintf(os.Stderr, "Usage: %s FUNCTION LOCATION\n\nExample: %[1]s current London,UK\n", os.Args[0])
		os.Exit(1)
	}

	location := GetLocation(os.Args)
	function := os.Args[1]
	c := NewClient(key)
	coordinates, err := c.GetCoordinates(location)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	conditions, forecast, err := c.GetWeather(coordinates)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	switch function {
	case FunctionCurrent:
		PrintCurrentConditions(conditions, forecast)
	case FunctionMoon:
		PrintMoon(forecast)
	default:
		fmt.Println()
		fmt.Println("This function isn't implemented yet. Please try it later again.")
		fmt.Println()
	}
	/*
		fmt.Println("Hours")
		for _, slot := range forecast.Hourly {
			fmt.Println(slot.Timestamp)
		}
		fmt.Println("Days")
		for _, slot := range forecast.Daily {
			fmt.Println(slot.Timestamp)
		}
	*/
}

func GetLocation(args []string) string {
	return strings.Join(args[2:], "+")
}

func GetFunction(args []string) string {
	return strings.Join(args[1:2], "")
}

func Get(location, key string) (Conditions, Forecast, error) {
	c := NewClient(key)
	coordinates, err := c.GetCoordinates(location)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	conditions, forecast, err := c.GetWeather(coordinates)
	if err != nil {
		return Conditions{}, Forecast{}, err
	}
	return conditions, forecast, nil
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

func ParseWeatherResponse(data []byte) (Conditions, Forecast, error) {
	var resp WeatherResponse
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return Conditions{}, Forecast{}, fmt.Errorf("invalid API response %s: %w", data, err)
	}
	if len(resp.Current.Weather) < 1 {
		return Conditions{}, Forecast{}, fmt.Errorf("invalid API response %s: want at least one Weather element", data)
	}
	conditions := Conditions{
		Timestamp:     time.Unix(resp.Current.DT, 0).Format("02.01.2006 15:04 MST"),
		Sunrise:       time.Unix(resp.Current.Sunrise, 0).Format("15:04"),
		Sunset:        time.Unix(resp.Current.Sunset, 0).Format("15:04"),
		Summary:       resp.Current.Weather[0].Description,
		Temperature:   resp.Current.Temp,
		FeelsLike:     resp.Current.Feels_Like,
		DewPoint:      resp.Current.Dew_Point,
		Pressure:      resp.Current.Pressure,
		Humidity:      resp.Current.Humidity,
		WindSpeed:     resp.Current.Wind_Speed,
		WindGust:      resp.Current.Wind_Gust,
		WindDirection: resp.Current.Wind_Deg,
	}
	forecast := Forecast{
		Hourly: []ForecastHourly{},
		Daily:  []ForecastDaily{},
	}
	for _, slot := range resp.Hourly {
		s := ForecastHourly{
			Timestamp: time.Unix(slot.DT, 0).Format("02.01.2006 15:04"),
		}
		forecast.Hourly = append(forecast.Hourly, s)
	}
	for _, slot := range resp.Daily {
		s := ForecastDaily{
			Timestamp: time.Unix(slot.DT, 0).Format("02.01.2006"),
			Moonrise:  time.Unix(slot.Moonrise, 0).Format("15:04"),
			Moonset:   time.Unix(slot.Moonset, 0).Format("15:04"),
			Moonphase: slot.Moon_Phase,
		}
		forecast.Daily = append(forecast.Daily, s)
	}
	return conditions, forecast, nil
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

// PrintCurrentConditions ... output of the current weather conditions, perfect if you can't look out of your window
func PrintCurrentConditions(conditions Conditions, forecast Forecast) {
	fmt.Println()
	fmt.Println(conditions.Timestamp)
	fmt.Println("-----------------------------------------------------")
	fmt.Printf("Sonne: %s / %s\n", conditions.Sunrise, conditions.Sunset)
	fmt.Printf("Mond: %s / %s, %s\n", forecast.Daily[0].Moonrise, forecast.Daily[0].Moonset, forecast.Daily[0].Moonphase.Description())
	fmt.Printf("Beschreibung: %s\n", conditions.Summary)
	fmt.Printf("Temperatur: %.1f °C, gefühlt %.1f °C\n", conditions.Temperature, conditions.FeelsLike)
	fmt.Printf("Taupunkt: %.1f °C\n", conditions.DewPoint)
	fmt.Printf("Luftdruck: %d hPa\n", conditions.Pressure)
	fmt.Printf("Luftfeuchtigkeit: %d %%\n", conditions.Humidity)
	fmt.Printf("Wind: %.0f km/h aus %s, in Böen %.0f km/h\n", conditions.WindSpeed.KmPerHour(), conditions.WindDirection.Direction(), conditions.WindGust.KmPerHour())
	fmt.Println()
}

// PrintMoon ... output of moonrise and moonset for next days, including the moon phases
func PrintMoon(forecast Forecast) {
	fmt.Println()
	fmt.Println("Mondauf-/untergang, Mondphase")
	fmt.Println("-----------------------------------------------------")
	lastDescription := ""
	for _, day := range forecast.Daily {
		currentDescritption := day.Moonphase.Description()
		if lastDescription != currentDescritption {
			fmt.Printf("%s: %s - %s, %s\n", day.Timestamp, day.Moonrise, day.Moonset, day.Moonphase.Description())
		} else {
			fmt.Printf("%s: %s - %s\n", day.Timestamp, day.Moonrise, day.Moonset)
		}
		lastDescription = currentDescritption
	}
	fmt.Println()
}

func GetTimestamp(sec int64, format string) string {
	return time.Unix(sec, 0).Format(format)
}

func (c *Client) FormatWeatherURL(coordinates Coordinates) string {
	return fmt.Sprintf("%s/data/3.0/onecall?lat=%g&lon=%g&units=metric&lang=de&appid=%s", c.BaseURL, coordinates.Lat, coordinates.Lon, c.APIKey)
}

func (c *Client) FormatGeoURL(location string) string {
	return fmt.Sprintf("%s/geo/1.0/direct?q=%s&limit=1&appid=%s", c.BaseURL, location, c.APIKey)
}

func (c *Client) GetWeather(coordinates Coordinates) (Conditions, Forecast, error) {
	URL := c.FormatWeatherURL(coordinates)
	resp, err := c.HTTPClient.Get(URL)
	if err != nil {
		return Conditions{}, Forecast{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Conditions{}, Forecast{}, fmt.Errorf("unexptected response status %q", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Conditions{}, Forecast{}, err
	}
	conditions, forecast, err := ParseWeatherResponse(data)
	if err != nil {
		return Conditions{}, Forecast{}, err
	}
	return conditions, forecast, nil
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

// KmPerHour ... helper method for speed output
func (s Speed) KmPerHour() float64 {
	return float64(s) * 3.6
}

// Direction ... converts degrees into human redable wind direction
func (d Direction) Direction() string {
	if (float64(d) > NNW+(360-NNW)/2 && float64(d) <= 360) || (float64(d) >= 0 && float64(d) <= NNO/2) {
		return "N"
	}
	if float64(d) > NNO/2 && float64(d) <= NNO+(NO-NNO)/2 {
		return "NNO"
	}
	if float64(d) > NNO+(NO-NNO)/2 && float64(d) <= NO+(ONO-NO)/2 {
		return "NO"
	}
	if float64(d) > NO+(ONO-NO)/2 && float64(d) <= ONO+(O-ONO)/2 {
		return "ONO"
	}
	if float64(d) > ONO+(O-ONO)/2 && float64(d) <= O+(OSO-O)/2 {
		return "O"
	}
	if float64(d) > O+(OSO-O)/2 && float64(d) <= OSO+(SO-OSO)/2 {
		return "OSO"
	}
	if float64(d) > OSO+(SO-OSO)/2 && float64(d) <= SO+(SSO-SO)/2 {
		return "SO"
	}
	if float64(d) > SO+(SSO-SO)/2 && float64(d) <= SSO+(S-SSO)/2 {
		return "SSO"
	}
	if float64(d) > SSO+(S-SSO)/2 && float64(d) <= S+(SSW-S)/2 {
		return "S"
	}
	if float64(d) > S+(SSW-S)/2 && float64(d) <= SSW+(SW-SSW)/2 {
		return "SSW"
	}
	if float64(d) > SSW+(SW-SSW)/2 && float64(d) <= SW+(WSW-SW)/2 {
		return "SW"
	}
	if float64(d) > SW+(WSW-SW)/2 && float64(d) <= WSW+(W-WSW)/2 {
		return "WSW"
	}
	if float64(d) > WSW+(W-WSW)/2 && float64(d) <= W+(WNW-W)/2 {
		return "W"
	}
	if float64(d) > W+(WNW-W)/2 && float64(d) <= WNW+(NW-WNW)/2 {
		return "WNW"
	}
	if float64(d) > WNW+(NW-WNW)/2 && float64(d) <= NW+(NNW-NW)/2 {
		return "NW"
	}
	if float64(d) > NW+(NNW-NW)/2 && float64(d) <= NNW+(360-NNW)/2 {
		return "NNW"
	}
	return "UNBEKANNT"
}

func (p Phase) Description() string {
	if float64(p) == 0 {
		return "Neumond"
	}
	if float64(p) > 0 && float64(p) < 0.25 {
		return "zunehmender Mond (vor Halbmond)"
	}
	if float64(p) == 0.25 {
		return "zunehmender Halbmond"
	}
	if float64(p) > 0.25 && float64(p) < 0.5 {
		return "zunehmender Mond (nach Halbmond)"
	}
	if float64(p) == 0.5 {
		return "Vollmond"
	}
	if float64(p) > 0.5 && float64(p) < 0.75 {
		return "abnehmender Mond (vor Halbmond)"
	}
	if float64(p) == 0.75 {
		return "abnehmender Halbmond"
	}
	if float64(p) > 0.75 && float64(p) < 1 {
		return "abnehmender Mond (nach Halbmond)"
	}
	if float64(p) == 1 {
		return "Neumond"
	}
	return "UNBEKANNT"
}
