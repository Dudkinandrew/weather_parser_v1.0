package main

import (
	"Weather_Parser/internal/config"
	"Weather_Parser/internal/geolocation"
	"Weather_Parser/internal/notify"
	"Weather_Parser/internal/storage"
	"Weather_Parser/internal/weather"
	"Weather_Parser/internal/web"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

// ============================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ============================================

func parseBool(value string, defaultValue bool) bool {
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes" || value == "True"
}

func parseDuration(value string, defaultValue time.Duration) time.Duration {
	if value == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		log.Printf("⚠️ Ошибка парсинга интервала '%s', используем значение по умолчанию", value)
		return defaultValue
	}
	return d
}

func maskToken(token string) string {
	if len(token) <= 10 {
		return "***"
	}
	return token[:10] + "..."
}

func containsStr(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func checkAlerts(data map[string]interface{}, prevTemp float64, prevCondition string) []string {
	temp := data["temp_c"].(float64)
	condition := data["condition"].(string)
	city := data["city"].(string)

	var alerts []string

	if prevTemp != 0 {
		diff := temp - prevTemp
		if diff > 5 {
			alerts = append(alerts, fmt.Sprintf(
				"🌡️ <b>Резкое потепление!</b>\n+%.1f°C за интервал\nБыло: %.1f°C → Стало: %.1f°C",
				diff, prevTemp, temp))
		} else if diff < -5 {
			alerts = append(alerts, fmt.Sprintf(
				"🌡️ <b>Резкое похолодание!</b>\n%.1f°C за интервал\nБыло: %.1f°C → Стало: %.1f°C",
				-diff, prevTemp, temp))
		}
	}

	if prevCondition != "" && prevCondition != condition {
		if containsStr(condition, "дожд") || containsStr(condition, "лив") {
			alerts = append(alerts, "☔ <b>Начался дождь!</b>\nНе забудьте зонт! ☂️")
		}
		if containsStr(condition, "снег") {
			alerts = append(alerts, "❄️ <b>Пошёл снег!</b>\nБудьте осторожны на дорогах! 🚗")
		}
		if containsStr(condition, "гроз") {
			alerts = append(alerts, "⚡ <b>Началась гроза!</b>\nОставайтесь в безопасности! 🏠")
		}
	}

	if temp > 35 {
		alerts = append(alerts, fmt.Sprintf(
			"☀️ <b>Сильная жара!</b>\n%.1f°C в %s\nПейте больше воды! 💧",
			temp, city))
	}
	if temp < -20 {
		alerts = append(alerts, fmt.Sprintf(
			"❄️ <b>Сильный мороз!</b>\n%.1f°C в %s\nОдевайтесь теплее! 🧣",
			temp, city))
	}
	if wind, ok := data["wind_kph"].(float64); ok && wind > 50 {
		alerts = append(alerts, fmt.Sprintf(
			"💨 <b>Сильный ветер!</b>\n%.1f км/ч в %s\nБудьте осторожны! 🌪️",
			wind, city))
	}

	return alerts
}

// ============================================
// СТРУКТУРА ДЛЯ ГОРОДОВ
// ============================================

type CityConfig struct {
	Name string
	Lat  float64
	Lon  float64
}

// ============================================
// ОСНОВНАЯ ФУНКЦИЯ
// ============================================

func main() {
	// ============================================
	// 1. ЗАГРУЗКА .env ФАЙЛА
	// ============================================
	if err := godotenv.Load(); err != nil {
		log.Println("ℹ️ .env файл не найден, используем флаги и значения по умолчанию")
	} else {
		log.Println("✅ .env файл загружен")
	}

	// ============================================
	// 2. ЧТЕНИЕ ПЕРЕМЕННЫХ ОКРУЖЕНИЯ
	// ============================================
	envCity := os.Getenv("CITY")
	envInterval := os.Getenv("INTERVAL")
	envAutoDetect := os.Getenv("AUTO_DETECT")
	envCSVPath := os.Getenv("CSV_PATH")

	envTelegramToken := os.Getenv("TELEGRAM_TOKEN")
	envTelegramChatID := os.Getenv("TELEGRAM_CHAT_ID")
	envTelegramEnabled := os.Getenv("TELEGRAM_ENABLED")

	envDBHost := os.Getenv("DB_HOST")
	envDBPort := os.Getenv("DB_PORT")
	envDBUser := os.Getenv("DB_USER")
	envDBPassword := os.Getenv("DB_PASSWORD")
	envDBName := os.Getenv("DB_NAME")
	envDBEnabled := os.Getenv("DB_ENABLED")

	// ============================================
	// 3. ПАРСИНГ ФЛАГОВ
	// ============================================
	city := flag.String("city", envCity, "Город для парсинга")
	autoDetect := flag.Bool("auto", parseBool(envAutoDetect, true), "Автоматическое определение города по IP")
	interval := flag.Duration("interval", parseDuration(envInterval, 30*time.Second), "Интервал обновления")
	csvPath := flag.String("csv", envCSVPath, "Путь к CSV файлу")

	telegramToken := flag.String("telegram-token", envTelegramToken, "Токен Telegram бота")
	telegramChatID := flag.String("telegram-chat", envTelegramChatID, "Chat ID для получения сообщений")
	telegramEnabled := flag.Bool("telegram-enabled", parseBool(envTelegramEnabled, false), "Включить Telegram бота")

	dbEnabled := flag.Bool("db-enabled", parseBool(envDBEnabled, false), "Использовать PostgreSQL")
	dbHost := flag.String("db-host", envDBHost, "Хост PostgreSQL")
	dbPort := flag.String("db-port", envDBPort, "Порт PostgreSQL")
	dbUser := flag.String("db-user", envDBUser, "Пользователь PostgreSQL")
	dbPassword := flag.String("db-password", envDBPassword, "Пароль PostgreSQL")
	dbName := flag.String("db-name", envDBName, "Имя базы данных PostgreSQL")

	flag.Parse()

	// ============================================
	// 4. СОЗДАНИЕ КОНФИГА
	// ============================================
	cfg := config.DefaultConfig()
	cfg.City = *city
	cfg.Interval = *interval
	cfg.CSVPath = *csvPath
	cfg.AutoDetect = *autoDetect
	cfg.TelegramEnabled = *telegramEnabled
	cfg.TelegramToken = *telegramToken
	cfg.TelegramChatID = *telegramChatID
	cfg.DBEnabled = *dbEnabled
	cfg.DBHost = *dbHost
	cfg.DBPort = *dbPort
	cfg.DBUser = *dbUser
	cfg.DBPassword = *dbPassword
	cfg.DBName = *dbName

	// ============================================
	// 5. ЛОГИРОВАНИЕ КОНФИГУРАЦИИ
	// ============================================
	log.Println("========================================")
	log.Println("📋 КОНФИГУРАЦИЯ:")
	log.Printf("   Город: %s", cfg.City)
	log.Printf("   Интервал: %v", cfg.Interval)
	log.Printf("   CSV файл: %s", cfg.CSVPath)
	log.Printf("   Автоопределение: %v", cfg.AutoDetect)
	log.Printf("   Telegram: %v", cfg.TelegramEnabled)
	if cfg.TelegramEnabled {
		log.Printf("   Telegram Token: %s...", maskToken(cfg.TelegramToken))
		log.Printf("   Telegram Chat ID: %s", cfg.TelegramChatID)
	}
	log.Printf("   PostgreSQL: %v", cfg.DBEnabled)
	if cfg.DBEnabled {
		log.Printf("   DB: %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)
	}
	log.Println("========================================")

	// ============================================
	// 6. ОПРЕДЕЛЕНИЕ МЕСТОПОЛОЖЕНИЯ (для основного города)
	// ============================================
	var lat, lon float64
	var cityName string

	if cfg.City != "" {
		coords, exists := weather.CityCoordinates[cfg.City]
		if !exists {
			log.Printf("⚠️ Город %s не найден в базе, используем Москву", cfg.City)
			coords = weather.CityCoordinates["Moscow"]
			cityName = "Moscow"
		} else {
			cityName = cfg.City
		}
		lat, lon = coords[0], coords[1]
		log.Printf("📍 Ручной режим: %s (%.4f, %.4f)", cityName, lat, lon)
	} else if cfg.AutoDetect {
		log.Println("📍 Определяем местоположение по IP...")
		city, latitude, longitude, err := geolocation.GetCityCoordinates()
		if err != nil {
			log.Printf("⚠️ Ошибка определения местоположения: %v", err)
			log.Println("ℹ️ Используем город по умолчанию: Москва")
			lat, lon = 55.7558, 37.6173
			cityName = "Moscow"
		} else {
			cityName = city
			lat = latitude
			lon = longitude
			log.Printf("✅ Определен город: %s (%.4f, %.4f)", cityName, lat, lon)
		}
	} else {
		log.Println("ℹ️ Используем город по умолчанию: Москва")
		lat, lon = 55.7558, 37.6173
		cityName = "Moscow"
	}
	cfg.City = cityName

	// ============================================
	// 7. СПИСОК ГОРОДОВ ДЛЯ ПАРСИНГА
	// ============================================
	citiesToParse := []CityConfig{
		{Name: "Волгоград", Lat: 48.7080, Lon: 44.5133},
		{Name: "Петров Вал", Lat: 50.1396, Lon: 45.2142},
		{Name: "Суводская", Lat: 49.4875, Lon: 45.0514},
		{Name: "Прохладный", Lat: 43.759, Lon: 44.011},
		{Name: "Терскол", Lat: 43.2573, Lon: 42.5082},
	}

	// ============================================
	// 8. ИНИЦИАЛИЗАЦИЯ ХРАНИЛИЩА
	// ============================================
	var storageClient interface {
		WriteWeather(data map[string]interface{}) error
	}

	if cfg.DBEnabled {
		pgStorage, err := storage.NewPostgresStorage(cfg.GetDBConnectionString())
		if err != nil {
			log.Printf("⚠️ Ошибка подключения к PostgreSQL: %v", err)
			log.Println("ℹ️ Используем CSV как запасной вариант")
			storageClient = storage.NewCSVStorage(cfg.CSVPath)
		} else {
			storageClient = pgStorage
			defer func() {
				if err := pgStorage.Close(); err != nil {
					log.Printf("⚠️ Ошибка закрытия БД: %v", err)
				} else {
					log.Println("✅ Соединение с БД закрыто")
				}
			}()
			log.Println("✅ Используем PostgreSQL")
		}
	} else {
		storageClient = storage.NewCSVStorage(cfg.CSVPath)
		log.Println("📁 Используем CSV")
	}

	// ============================================
	// 9. ЗАПУСК ВЕБ-СЕРВЕРА
	// ============================================
	var webStorage web.StorageReader
	if pgStorage, ok := storageClient.(*storage.PostgresStorage); ok {
		webStorage = pgStorage
	} else if csvStorage, ok := storageClient.(*storage.CSVStorage); ok {
		webStorage = csvStorage
	}

	if webStorage != nil {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8081"
		}
		webServer := web.NewServer(":"+port, webStorage)
		go func() {
			if err := webServer.Start(); err != nil {
				log.Printf("⚠️ Ошибка веб-сервера: %v", err)
			}
		}()
		log.Printf("🌐 Веб-интерфейс запущен на http://localhost:%s", port)
	} else {
		log.Println("⚠️ Хранилище не поддерживает веб-интерфейс")
	}

	// ============================================
	// 10. ИНИЦИАЛИЗАЦИЯ TELEGRAM БОТА
	// ============================================
	var telegramBot *notify.TelegramBot
	if cfg.TelegramEnabled && cfg.TelegramToken != "" && cfg.TelegramChatID != "" {
		telegramBot = notify.NewTelegramBot(cfg.TelegramToken, cfg.TelegramChatID)
		if err := telegramBot.SendTestMessage(); err != nil {
			log.Printf("⚠️ Не удалось отправить приветствие в Telegram: %v", err)
			log.Println("💡 Проверьте интернет, VPN или прокси")
		} else {
			log.Println("✅ Приветственное сообщение отправлено в Telegram")
		}
	} else if cfg.TelegramEnabled {
		log.Println("⚠️ Telegram включен, но не указан токен или Chat ID")
	}

	// ============================================
	// 11. ИНИЦИАЛИЗАЦИЯ КЛИЕНТА ПОГОДЫ
	// ============================================
	weatherClient := weather.NewOpenMeteoClient()

	log.Printf("🚀 Запуск парсера погоды для %d городов", len(citiesToParse))
	log.Printf("📁 Данные будут сохранены в: %s", cfg.CSVPath)
	log.Printf("⏱️ Интервал обновления: %v", cfg.Interval)

	// ============================================
	// 12. ФУНКЦИЯ ПАРСИНГА ВСЕХ ГОРОДОВ
	// ============================================
	fetchAllCities := func() {
		log.Printf("📡 Запрашиваем погоду для %d городов...", len(citiesToParse))

		for _, city := range citiesToParse {
			data, err := weatherClient.GetCurrentWeather(city.Lat, city.Lon)
			if err != nil {
				log.Printf("   ❌ Ошибка для %s: %v", city.Name, err)
				continue
			}

			data["city"] = city.Name

			temp := data["temp_c"].(float64)
			condition := data["condition"].(string)
			log.Printf("   🌍 %s: %.1f°C, %s", city.Name, temp, condition)

			if err := storageClient.WriteWeather(data); err != nil {
				log.Printf("   ❌ Ошибка сохранения для %s: %v", city.Name, err)
			}
		}

		log.Println("✅ Парсинг всех городов завершен")
	}

	// Первый запуск
	fetchAllCities()

	// Таймер для периодического выполнения
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	// Канал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем горутину для обработки сигналов
	go func() {
		for {
			select {
			case <-ticker.C:
				fetchAllCities()
			case <-quit:
				if cfg.TelegramEnabled && telegramBot != nil {
					if err := telegramBot.SendMessage("🛑 <b>Погодный монитор остановлен</b>"); err != nil {
						log.Printf("⚠️ Ошибка отправки сообщения о остановке: %v", err)
					} else {
						log.Println("📱 Сообщение о остановке отправлено в Telegram")
					}
				}
				log.Println("👋 Завершаем работу...")
				return
			}
		}
	}()

	log.Println("🔄 Программа работает, нажмите Ctrl+C для остановки")
	log.Println("")
	log.Println("💡 Способы запуска:")
	log.Println("  1. С .env файлом:          go run .")
	log.Println("  2. С флагами:              go run . -city=Volgograd -interval=1m")
	log.Println("  3. С Telegram:             go run . -telegram-token=TOKEN -telegram-chat=ID")
	log.Println("  4. С PostgreSQL:           go run . -db-enabled=true -db-host=localhost")
	log.Println("  5. Веб-интерфейс:          http://localhost:8081")
	log.Println("")

	// Блокируем main горутину до получения сигнала
	<-quit
}
