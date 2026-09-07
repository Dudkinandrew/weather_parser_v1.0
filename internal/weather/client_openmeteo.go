package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenMeteoData — структура для Open-Meteo API
type OpenMeteoData struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
		WindSpeed   float64 `json:"windspeed"`
		WeatherCode int     `json:"weathercode"`
		Time        string  `json:"time"`
	} `json:"current_weather"`
	Daily struct {
		TemperatureMax []float64 `json:"temperature_2m_max"`
		TemperatureMin []float64 `json:"temperature_2m_min"`
		Time           []string  `json:"time"`
	} `json:"daily"`
}

type OpenMeteoClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewOpenMeteoClient() *OpenMeteoClient {
	return &OpenMeteoClient{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		baseURL:    "https://api.open-meteo.com/v1",
	}
}

// GetCurrentWeather получает погоду по координатам
func (c *OpenMeteoClient) GetCurrentWeather(lat, lon float64) (map[string]interface{}, error) {
	// Open-Meteo использует координаты, не города
	url := fmt.Sprintf(
		"%s/forecast?latitude=%.4f&longitude=%.4f&current_weather=true&daily=temperature_2m_max,temperature_2m_min&timezone=auto",
		c.baseURL, lat, lon,
	)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API вернул статус: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения: %w", err)
	}

	var data OpenMeteoData
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("ошибка парсинга: %w", err)
	}

	// Коды погоды: https://open-meteo.com/en/docs
	weatherCodes := map[int]string{
		0:  "Ясно",
		1:  "Преимущественно ясно",
		2:  "Переменная облачность",
		3:  "Пасмурно",
		45: "Туман",
		51: "Морось (слабая)",
		61: "Дождь (слабый)",
		71: "Снег (слабый)",
		80: "Ливень",
		95: "Гроза",
	}

	condition := weatherCodes[data.CurrentWeather.WeatherCode]
	if condition == "" {
		condition = fmt.Sprintf("Код: %d", data.CurrentWeather.WeatherCode)
	}

	// Проверяем, есть ли прогноз на сегодня
	var maxTemp, minTemp float64
	if len(data.Daily.TemperatureMax) > 0 {
		maxTemp = data.Daily.TemperatureMax[0]
	}
	if len(data.Daily.TemperatureMin) > 0 {
		minTemp = data.Daily.TemperatureMin[0]
	}

	result := map[string]interface{}{
		"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
		"city":       fmt.Sprintf("Координаты: %.4f, %.4f", lat, lon),
		"country":    "",
		"temp_c":     data.CurrentWeather.Temperature,
		"humidity":   0, // Open-Meteo не дает влажность в бесплатном API
		"condition":  condition,
		"wind_kph":   data.CurrentWeather.WindSpeed * 3.6, // м/с -> км/ч
		"max_temp_c": maxTemp,
		"min_temp_c": minTemp,
	}

	return result, nil
}

// Координаты популярных городов
var CityCoordinates = map[string][2]float64{
	"Moscow":    {55.7558, 37.6173},
	"London":    {51.5074, -0.1278},
	"New York":  {40.7128, -74.0060},
	"Tokyo":     {35.6762, 139.6503},
	"Berlin":    {52.5200, 13.4050},
	"Paris":     {48.8566, 2.3522},
	"Volgograd": {48.7080, 44.5133},
}
