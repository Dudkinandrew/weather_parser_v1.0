package config

import "time"

type Config struct {
	// Основные настройки
	City       string
	Latitude   float64
	Longitude  float64
	Interval   time.Duration
	CSVPath    string
	AutoDetect bool

	// Telegram
	TelegramToken   string
	TelegramChatID  string
	TelegramEnabled bool

	// PostgreSQL
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBEnabled  bool
}

func DefaultConfig() Config {
	return Config{
		City:       "Moscow",
		Interval:   30 * time.Second,
		CSVPath:    "weather_data.csv",
		AutoDetect: true,

		TelegramEnabled: true,
		TelegramToken:   "",
		TelegramChatID:  "",

		DBEnabled:  false,
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "postgres",
		DBPassword: "postgres",
		DBName:     "weather",
	}
}

// GetDBConnectionString возвращает строку подключения к БД
func (c *Config) GetDBConnectionString() string {
	return "host=" + c.DBHost +
		" port=" + c.DBPort +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=disable"
}
