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
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
	Current   struct {
		Time             string  `json:"time"`
		Temperature      float64 `json:"temperature_2m"`
		WindSpeed        float64 `json:"wind_speed_10m"`
		WeatherCode      int     `json:"weather_code"`
		RelativeHumidity int     `json:"relative_humidity_2m"` // 👈 ВЛАЖНОСТЬ
	} `json:"current"`
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

// GetCurrentWeather получает погоду по координатам (с влажностью!)
func (c *OpenMeteoClient) GetCurrentWeather(lat, lon float64) (map[string]interface{}, error) {
	// Явно запрашиваем влажность в параметре current
	url := fmt.Sprintf(
		"%s/forecast?latitude=%.4f&longitude=%.4f&current=temperature_2m,relative_humidity_2m,wind_speed_10m,weather_code&daily=temperature_2m_max,temperature_2m_min&timezone=auto",
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

	// Коды погоды
	weatherCodes := map[int]string{
		0:  "Ясно",
		1:  "Преимущественно ясно",
		2:  "Переменная облачность",
		3:  "Пасмурно",
		45: "Туман",
		48: "Туман",
		51: "Морось (слабая)",
		53: "Морось (умеренная)",
		55: "Морось (сильная)",
		56: "Морось (замерзающая)",
		57: "Морось (замерзающая)",
		61: "Дождь (слабый)",
		63: "Дождь (умеренный)",
		65: "Дождь (сильный)",
		71: "Снег (слабый)",
		73: "Снег (умеренный)",
		75: "Снег (сильный)",
		80: "Ливень (слабый)",
		81: "Ливень (умеренный)",
		82: "Ливень (сильный)",
		95: "Гроза",
		96: "Гроза с градом",
		99: "Гроза с градом",
	}

	condition := weatherCodes[data.Current.WeatherCode]
	if condition == "" {
		condition = fmt.Sprintf("Код: %d", data.Current.WeatherCode)
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
		"temp_c":     data.Current.Temperature,
		"humidity":   data.Current.RelativeHumidity, // 👈 ВЛАЖНОСТЬ ПОЯВИЛАСЬ
		"condition":  condition,
		"wind_kph":   data.Current.WindSpeed * 3.6, // м/с → км/ч
		"max_temp_c": maxTemp,
		"min_temp_c": minTemp,
	}

	return result, nil
}

// Координаты популярных городов
var CityCoordinates = map[string][2]float64{
	"Moscow":           {55.7558, 37.6173},
	"Volgograd":        {48.7080, 44.5133},
	"London":           {51.5074, -0.1278},
	"New York":         {40.7128, -74.0060},
	"Tokyo":            {35.6762, 139.6503},
	"Berlin":           {52.5200, 13.4050},
	"Paris":            {48.8566, 2.3522},
	"Saint Petersburg": {59.9343, 30.3351},
	"Sochi":            {43.5855, 39.7231},
	"Kazan":            {55.8304, 49.0661},
	"Novosibirsk":      {55.0084, 82.9357},
	"Yekaterinburg":    {56.8389, 60.6057},
}
