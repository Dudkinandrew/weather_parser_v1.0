package geolocation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// IPInfo — структура ответа от сервиса ip-api.com
type IPInfo struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	As          string  `json:"as"`
	Query       string  `json:"query"`
}

// GetLocationByIP определяет местоположение по IP-адресу
func GetLocationByIP() (*IPInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	// Бесплатный сервис, не требует ключа
	// Используем ip-api.com с JSON форматом
	resp, err := client.Get("http://ip-api.com/json/?fields=status,country,countryCode,region,regionName,city,zip,lat,lon,timezone,isp,org,as,query")
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к геолокации: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	var info IPInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("ошибка парсинга геолокации: %w", err)
	}

	if info.Status != "success" {
		return nil, fmt.Errorf("геолокация не удалась: статус %s", info.Status)
	}

	return &info, nil
}

// GetCityCoordinates возвращает координаты города из IP
func GetCityCoordinates() (city string, lat, lon float64, err error) {
	info, err := GetLocationByIP()
	if err != nil {
		return "", 0, 0, err
	}

	return info.City, info.Lat, info.Lon, nil
}
