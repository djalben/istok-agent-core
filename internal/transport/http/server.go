package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/istok/agent-core/internal/application/usecases"
)

// Server - HTTP сервер
type Server struct {
	addr             string
	projectGenerator *usecases.ProjectGeneratorService
	server           *http.Server
}

// NewServer создает новый HTTP сервер
func NewServer(addr string, projectGenerator *usecases.ProjectGeneratorService) *Server {
	return &Server{
		addr:             addr,
		projectGenerator: projectGenerator,
	}
}

// Start запускает HTTP сервер
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Регистрация handlers
	generateHandler := NewGenerateHandler(s.projectGenerator)
	statsHandler := NewStatsHandler(s.projectGenerator)
	healthHandler := NewHealthHandler()

	// API endpoints
	mux.HandleFunc("/api/v1/generate", s.corsMiddleware(generateHandler.Handle))
	mux.HandleFunc("/api/v1/stats", s.corsMiddleware(statsHandler.Handle))
	mux.HandleFunc("/api/v1/health", s.corsMiddleware(healthHandler.Handle))

	// Middleware chain
	handler := s.loggingMiddleware(mux)

	s.server = &http.Server{
		Addr:         s.addr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🚀 HTTP сервер запущен на %s\n", s.addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully останавливает сервер
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("⏳ Остановка HTTP сервера...")
	return s.server.Shutdown(ctx)
}

// corsMiddleware добавляет CORS headers
func (s *Server) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// CORS headers для доступа с localhost:3000
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Обработка preflight запросов
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// loggingMiddleware логирует все запросы
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем wrapper для ResponseWriter чтобы захватить status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf(
			"📝 %s %s - %d (%v)",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
		)
	})
}

// responseWriter оборачивает http.ResponseWriter для захвата status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// writeJSON отправляет JSON ответ
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Для простоты используем fmt.Fprintf, в production лучше encoding/json
	_, err := fmt.Fprintf(w, "%v", data)
	return err
}

// writeError отправляет JSON ошибку
func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"error": "%s"}`, message)
}
