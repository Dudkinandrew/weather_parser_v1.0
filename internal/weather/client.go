package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WeatherData — структура ответа от API
type WeatherData struct {
	Current struct {
		TempC     float64 `json:"temp_c"`
		Humidity  int     `json:"humidity"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
		WindKph float64 `json:"wind_kph"`
	} `json:"current"`
	Location struct {
		Name    string `json:"name"`
		Country string `json:"country"`
	} `json:"location"`
	Forecast struct {
		ForecastDay []struct {
			Date string `json:"date"`
			Day  struct {
				MaxTempC float64 `json:"maxtemp_c"`
				MinTempC float64 `json:"mintemp_c"`
			} `json:"day"`
		} `json:"forecastday"`
	} `json:"forecast"`
}

// Client — клиент для работы с погодным API
type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://wttr.in",
	}
}

// FetchWeather получает погоду для указанного города
func (c *Client) FetchWeather(city string) (*WeatherData, error) {
	// Формируем URL: ?format=j1 означает JSON
	url := fmt.Sprintf("%s/%s?format=j1&lang=ru", c.baseURL, city)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Добавляем User-Agent, чтобы некоторые API не банили
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API вернул статус: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	var data WeatherData
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	return &data, nil
}

// GetCurrentWeather возвращает только нужные поля (для удобства)
func (c *Client) GetCurrentWeather(city string) (map[string]interface{}, error) {
	data, err := c.FetchWeather(city)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
		"city":       data.Location.Name,
		"country":    data.Location.Country,
		"temp_c":     data.Current.TempC,
		"humidity":   data.Current.Humidity,
		"condition":  data.Current.Condition.Text,
		"wind_kph":   data.Current.WindKph,
		"max_temp_c": data.Forecast.ForecastDay[0].Day.MaxTempC,
		"min_temp_c": data.Forecast.ForecastDay[0].Day.MinTempC,
	}

	return result, nil
}
