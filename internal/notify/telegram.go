package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// TelegramBot — клиент для отправки сообщений в Telegram
type TelegramBot struct {
	token    string
	chatID   string
	client   *http.Client
	baseURL  string
	maxRetry int
}

// NewTelegramBot создает нового бота
func NewTelegramBot(token, chatID string) *TelegramBot {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Настройка прокси из переменных окружения
	if proxyURL := getProxyFromEnv(); proxyURL != nil {
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		fmt.Println("🔌 Используется прокси для Telegram")
	}

	return &TelegramBot{
		token:    token,
		chatID:   chatID,
		client:   client,
		baseURL:  "https://api.telegram.org",
		maxRetry: 3,
	}
}

// SendMessage отправляет текстовое сообщение
func (t *TelegramBot) SendMessage(text string) error {
	url := fmt.Sprintf("%s/bot%s/sendMessage", t.baseURL, t.token)

	payload := map[string]interface{}{
		"chat_id":    t.chatID,
		"text":       text,
		"parse_mode": "HTML",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}

	var lastErr error
	for i := 0; i < t.maxRetry; i++ {
		if i > 0 {
			waitTime := time.Duration(i*2) * time.Second
			fmt.Printf("🔄 Повторная попытка %d через %v...\n", i+1, waitTime)
			time.Sleep(waitTime)
		}

		resp, err := t.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = fmt.Errorf("ошибка отправки: %w", err)
			continue
		}

		// Читаем тело ответа для диагностики
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("API вернул статус %d: %s", resp.StatusCode, string(body))
			continue
		}

		return nil // успех
	}

	return fmt.Errorf("не удалось после %d попыток: %w", t.maxRetry, lastErr)
}

// SendWeather отправляет красивое сообщение с погодой
func (t *TelegramBot) SendWeather(data map[string]interface{}) error {
	temp, _ := data["temp_c"].(float64)
	condition, _ := data["condition"].(string)
	city, _ := data["city"].(string)
	humidity, _ := data["humidity"].(int)
	wind, _ := data["wind_kph"].(float64)
	timestamp, _ := data["timestamp"].(string)

	tempIcon := getTempIcon(temp)
	weatherIcon := getWeatherEmoji(condition)

	message := fmt.Sprintf(
		`<b>🌍 Погода в %s</b>
%s <b>%.1f°C</b>
%s <b>%s</b>
💧 Влажность: <b>%d%%</b>
💨 Ветер: <b>%.1f м/с</b>
🕐 %s

#погода #%s`,
		city,
		weatherIcon, temp,
		tempIcon, condition,
		humidity,
		wind,
		timestamp,
		city,
	)

	return t.SendMessage(message)
}

// SendAlert отправляет сообщение-предупреждение
func (t *TelegramBot) SendAlert(alert string) error {
	message := fmt.Sprintf("⚠️ <b>ВНИМАНИЕ!</b>\n%s", alert)
	return t.SendMessage(message)
}

// SendTestMessage отправляет тестовое сообщение для проверки
func (t *TelegramBot) SendTestMessage() error {
	return t.SendMessage("✅ <b>Бот работает!</b>\nПодключение к Telegram API успешно установлено.")
}

// getProxyFromEnv получает прокси из переменных окружения
func getProxyFromEnv() *url.URL {
	proxyEnv := []string{"HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy"}

	for _, env := range proxyEnv {
		if proxyStr := os.Getenv(env); proxyStr != "" {
			if proxyURL, err := url.Parse(proxyStr); err == nil {
				return proxyURL
			}
		}
	}
	return nil
}

// getTempIcon возвращает иконку для температуры
func getTempIcon(temp float64) string {
	if temp > 30 {
		return "🔥"
	} else if temp > 20 {
		return "☀️"
	} else if temp > 10 {
		return "🌤️"
	} else if temp > 0 {
		return "🌥️"
	}
	return "❄️"
}

// getWeatherEmoji возвращает эмодзи для погоды
func getWeatherEmoji(condition string) string {
	switch {
	case containsStr(condition, "ясно") || containsStr(condition, "солнеч"):
		return "☀️"
	case containsStr(condition, "облач"):
		return "⛅"
	case containsStr(condition, "пасмур"):
		return "☁️"
	case containsStr(condition, "дожд") || containsStr(condition, "лив"):
		return "🌧️"
	case containsStr(condition, "снег"):
		return "❄️"
	case containsStr(condition, "туман"):
		return "🌫️"
	case containsStr(condition, "гроз"):
		return "⛈️"
	case containsStr(condition, "ветр"):
		return "💨"
	default:
		return "🌤️"
	}
}

// containsStr проверяет, содержится ли подстрока
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
