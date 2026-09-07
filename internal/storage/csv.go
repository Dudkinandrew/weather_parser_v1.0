package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

type CSVStorage struct {
	filePath string
	mu       sync.Mutex // защищаем файл от конкурентной записи
}

func NewCSVStorage(filePath string) *CSVStorage {
	return &CSVStorage{
		filePath: filePath,
	}
}

func (s *CSVStorage) GetLatest(city string) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, nil
	}

	// Идем с конца — ищем последнюю запись для указанного города
	for i := len(records) - 1; i >= 1; i-- {
		record := records[i]
		if len(record) < 9 {
			continue
		}
		if record[1] == city { // город во второй колонке
			temp, _ := strconv.ParseFloat(record[3], 64)
			humidity, _ := strconv.Atoi(record[4])
			wind, _ := strconv.ParseFloat(record[6], 64)
			maxTemp, _ := strconv.ParseFloat(record[7], 64)
			minTemp, _ := strconv.ParseFloat(record[8], 64)

			return map[string]interface{}{
				"timestamp": record[0],
				"city":      record[1],
				"country":   record[2],
				"temp_c":    temp,
				"humidity":  humidity,
				"condition": record[5],
				"wind_kph":  wind,
				"max_temp":  maxTemp,
				"min_temp":  minTemp,
			}, nil
		}
	}
	return nil, nil // город не найден
}

// GetHistory возвращает историю из CSV
func (s *CSVStorage) GetHistory(city string, days int) ([]map[string]interface{}, error) {
	file, err := os.Open(s.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return []map[string]interface{}{}, nil
	}

	// Парсим заголовки

	var history []map[string]interface{}
	cutoff := time.Now().AddDate(0, 0, -days)

	for _, record := range records[1:] {
		if len(record) < 6 {
			continue
		}

		timestamp, err := time.Parse("2006-01-02 15:04:05", record[0])
		if err != nil {
			continue
		}

		if timestamp.Before(cutoff) {
			continue
		}

		temp, _ := strconv.ParseFloat(record[3], 64)
		humidity, _ := strconv.Atoi(record[4])

		history = append(history, map[string]interface{}{
			"timestamp": timestamp.Format("2006-01-02 15:04"),
			"temp_c":    temp,
			"humidity":  humidity,
			"condition": record[5],
		})
	}

	return history, nil
}

// GetStats возвращает статистику за последние N дней для города
func (s *CSVStorage) GetStats(city string, days int) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return map[string]interface{}{
			"city":          city,
			"days":          days,
			"avg_temp":      0.0,
			"max_temp":      0.0,
			"min_temp":      0.0,
			"avg_humidity":  0.0,
			"total_records": 0,
		}, nil
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	var temps []float64
	var humidities []int

	for _, record := range records[1:] {
		if len(record) < 6 {
			continue
		}
		if record[1] != city {
			continue
		}
		timestamp, err := time.Parse("2006-01-02 15:04:05", record[0])
		if err != nil {
			continue
		}
		if timestamp.Before(cutoff) {
			continue
		}
		temp, _ := strconv.ParseFloat(record[3], 64)
		humidity, _ := strconv.Atoi(record[4])
		temps = append(temps, temp)
		humidities = append(humidities, humidity)
	}

	if len(temps) == 0 {
		return map[string]interface{}{
			"city":          city,
			"days":          days,
			"avg_temp":      0.0,
			"max_temp":      0.0,
			"min_temp":      0.0,
			"avg_humidity":  0.0,
			"total_records": 0,
		}, nil
	}

	// Вычисляем статистику
	var sumTemp, sumHum float64
	maxTemp := temps[0]
	minTemp := temps[0]
	for _, t := range temps {
		sumTemp += t
		if t > maxTemp {
			maxTemp = t
		}
		if t < minTemp {
			minTemp = t
		}
	}
	for _, h := range humidities {
		sumHum += float64(h)
	}
	avgTemp := sumTemp / float64(len(temps))
	avgHumidity := sumHum / float64(len(humidities))

	return map[string]interface{}{
		"city":          city,
		"days":          days,
		"avg_temp":      avgTemp,
		"max_temp":      maxTemp,
		"min_temp":      minTemp,
		"avg_humidity":  avgHumidity,
		"total_records": len(temps),
	}, nil
}

// WriteWeather записывает данные погоды в CSV
func (s *CSVStorage) WriteWeather(data map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем, существует ли файл
	fileExists := true
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		fileExists = false
	}

	// Открываем файл для добавления
	file, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Если файл новый — пишем заголовки
	if !fileExists {
		headers := []string{
			"timestamp", "city", "country", "temp_c", "humidity",
			"condition", "wind_kph", "max_temp_c", "min_temp_c",
		}
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("ошибка записи заголовков: %w", err)
		}
	}

	// Формируем строку данных в правильном порядке
	row := []string{
		data["timestamp"].(string),
		data["city"].(string),
		data["country"].(string),
		fmt.Sprintf("%.1f", data["temp_c"]),
		fmt.Sprintf("%d", data["humidity"]),
		data["condition"].(string),
		fmt.Sprintf("%.1f", data["wind_kph"]),
		fmt.Sprintf("%.1f", data["max_temp_c"]),
		fmt.Sprintf("%.1f", data["min_temp_c"]),
	}

	if err := writer.Write(row); err != nil {
		return fmt.Errorf("ошибка записи данных: %w", err)
	}

	return nil
}
