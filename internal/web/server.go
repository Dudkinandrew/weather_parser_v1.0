package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)

// ============================================
// ОСНОВНАЯ СТРУКТУРА СЕРВЕРА
// ============================================

type Server struct {
	port    string
	storage StorageReader
}

type StorageReader interface {
	GetLatest(city string) (map[string]interface{}, error)
	GetStats(city string, days int) (map[string]interface{}, error)
	GetHistory(city string, days int) ([]map[string]interface{}, error)
}

// ============================================
// СОЗДАНИЕ СЕРВЕРА
// ============================================

func NewServer(port string, storage StorageReader) *Server {
	return &Server{
		port:    port,
		storage: storage,
	}
}

// ============================================
// ЗАПУСК СЕРВЕРА
// ============================================

func (s *Server) Start() error {
	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/api/weather/latest", s.handleLatest)
	http.HandleFunc("/api/weather/stats", s.handleStats)
	http.HandleFunc("/api/weather/history", s.handleHistory)
	http.HandleFunc("/api/weather/update", s.handleUpdate)

	log.Printf("🌐 Веб-интерфейс запущен на http://localhost%s", s.port)
	return http.ListenAndServe(s.port, nil)
}

// ============================================
// ОБРАБОТЧИКИ
// ============================================

// handleIndex — главная страница (теперь загружает шаблон из файла)
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	// Определяем путь к шаблону
	tmplPath := filepath.Join("templates", "index.html")

	// Парсим шаблон из файла
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка загрузки шаблона: %v", err), http.StatusInternalServerError)
		return
	}

	// Выполняем шаблон
	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка выполнения шаблона: %v", err), http.StatusInternalServerError)
		return
	}
}

// handleLatest — возвращает последние данные
func (s *Server) handleLatest(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	if city == "" {
		city = "Volgograd"
	}

	data, err := s.storage.GetLatest(city)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка получения данных: %v", err), http.StatusInternalServerError)
		return
	}

	if data == nil {
		http.Error(w, "Нет данных для этого города", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// handleStats — возвращает статистику
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	if city == "" {
		city = "Volgograd"
	}

	days := 7
	if d := r.URL.Query().Get("days"); d != "" {
		if val, err := strconv.Atoi(d); err == nil && val > 0 {
			days = val
		}
	}

	stats, err := s.storage.GetStats(city, days)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка получения статистики: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleHistory — возвращает историю для графика
func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	if city == "" {
		city = "Volgograd"
	}

	days := 7
	if d := r.URL.Query().Get("days"); d != "" {
		if val, err := strconv.Atoi(d); err == nil && val > 0 {
			days = val
		}
	}

	history, err := s.storage.GetHistory(city, days)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка получения истории: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// handleUpdate — принудительное обновление (заглушка)
func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Обновление запущено",
	})
}
