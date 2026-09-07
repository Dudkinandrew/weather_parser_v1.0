package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// PostgresStorage — хранилище в PostgreSQL
type PostgresStorage struct {
	db *sql.DB
}

// GetHistory возвращает историю за последние N дней
func (s *PostgresStorage) GetHistory(city string, days int) ([]map[string]interface{}, error) {
	query := `
	SELECT timestamp, temp_c, humidity, condition
	FROM weather 
	WHERE city = $1 AND timestamp > NOW() - INTERVAL '%d days'
	ORDER BY timestamp ASC
	`

	query = fmt.Sprintf(query, days)

	rows, err := s.db.Query(query, city)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения истории: %w", err)
	}
	defer rows.Close()

	var history []map[string]interface{}

	for rows.Next() {
		var timestamp time.Time
		var temp float64
		var humidity int
		var condition string

		if err := rows.Scan(&timestamp, &temp, &humidity, &condition); err != nil {
			continue
		}

		history = append(history, map[string]interface{}{
			"timestamp": timestamp.Format("2006-01-02 15:04"),
			"temp_c":    temp,
			"humidity":  humidity,
			"condition": condition,
		})
	}

	return history, nil
}

// NewPostgresStorage создает подключение к PostgreSQL
func NewPostgresStorage(connString string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("БД не отвечает: %w", err)
	}

	// Создаем таблицу, если её нет
	if err := createTable(db); err != nil {
		return nil, fmt.Errorf("ошибка создания таблицы: %w", err)
	}

	log.Println("✅ PostgreSQL подключен")
	return &PostgresStorage{db: db}, nil
}

// createTable создает таблицу для погоды
func createTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS weather (
		id SERIAL PRIMARY KEY,
		timestamp TIMESTAMP NOT NULL,
		city VARCHAR(100) NOT NULL,
		country VARCHAR(100),
		temp_c DECIMAL(5,2) NOT NULL,
		humidity INT,
		condition VARCHAR(100),
		wind_kph DECIMAL(5,2),
		max_temp_c DECIMAL(5,2),
		min_temp_c DECIMAL(5,2),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_weather_timestamp ON weather(timestamp);
	CREATE INDEX IF NOT EXISTS idx_weather_city ON weather(city);
	`

	_, err := db.Exec(query)
	return err
}

// WriteWeather сохраняет данные в PostgreSQL
func (s *PostgresStorage) WriteWeather(data map[string]interface{}) error {
	query := `
	INSERT INTO weather (
		timestamp, city, country, temp_c, humidity, 
		condition, wind_kph, max_temp_c, min_temp_c
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	// Парсим timestamp
	timestamp, err := time.Parse("2006-01-02 15:04:05", data["timestamp"].(string))
	if err != nil {
		// Если не получилось, используем текущее время
		timestamp = time.Now()
	}

	_, err = s.db.Exec(query,
		timestamp,
		data["city"],
		data["country"],
		data["temp_c"],
		data["humidity"],
		data["condition"],
		data["wind_kph"],
		data["max_temp_c"],
		data["min_temp_c"],
	)

	if err != nil {
		return fmt.Errorf("ошибка записи в БД: %w", err)
	}

	return nil
}

// GetLatest возвращает последнюю запись для города
func (s *PostgresStorage) GetLatest(city string) (map[string]interface{}, error) {
	query := `
	SELECT timestamp, city, country, temp_c, humidity, condition, wind_kph
	FROM weather 
	WHERE city = $1 
	ORDER BY timestamp DESC 
	LIMIT 1
	`

	var timestamp time.Time
	var temp, wind float64
	var humidity int
	var cityName, country, condition string

	err := s.db.QueryRow(query, city).Scan(
		&timestamp, &cityName, &country, &temp, &humidity, &condition, &wind,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Нет данных
		}
		return nil, fmt.Errorf("ошибка получения данных: %w", err)
	}

	result := map[string]interface{}{
		"timestamp": timestamp.Format("2006-01-02 15:04:05"),
		"city":      cityName,
		"country":   country,
		"temp_c":    temp,
		"humidity":  humidity,
		"condition": condition,
		"wind_kph":  wind,
	}

	return result, nil
}

// GetStats возвращает статистику за последние N дней
func (s *PostgresStorage) GetStats(city string, days int) (map[string]interface{}, error) {
	query := `
	SELECT 
		AVG(temp_c) as avg_temp,
		MAX(temp_c) as max_temp,
		MIN(temp_c) as min_temp,
		AVG(humidity) as avg_humidity,
		COUNT(*) as records
	FROM weather 
	WHERE city = $1 AND timestamp > NOW() - INTERVAL '%d days'
	`

	query = fmt.Sprintf(query, days)

	var avgTemp, maxTemp, minTemp, avgHumidity float64
	var records int

	err := s.db.QueryRow(query, city).Scan(
		&avgTemp, &maxTemp, &minTemp, &avgHumidity, &records,
	)

	if err != nil {
		return nil, fmt.Errorf("ошибка статистики: %w", err)
	}

	return map[string]interface{}{
		"city":          city,
		"days":          days,
		"avg_temp":      avgTemp,
		"max_temp":      maxTemp,
		"min_temp":      minTemp,
		"avg_humidity":  avgHumidity,
		"total_records": records,
	}, nil
}

// Close закрывает соединение с БД
func (s *PostgresStorage) Close() error {
	return s.db.Close()

}
