package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

// HTTPError представляет ошибку HTTP с кодом и сообщением.
type HTTPError struct {
	Status  int    `json:"-"`
	Message string `json:"error"`
}

func (e *HTTPError) Error() string {
	return e.Message
}

// newHTTPError создаёт новый экземпляр HTTPError.
func newHTTPError(status int, msg string) *HTTPError {
	return &HTTPError{Status: status, Message: msg}
}

// LoggingMiddleware логирует начало и конец обработки каждого запроса с временем выполнения.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[START] %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Printf("[END] %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// RecoveryMiddleware перехватывает паники и возвращает статус 500.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[PANIC] %v", rec)
				writeJSON(w, http.StatusInternalServerError, HTTPError{Message: "internal server error"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// HandlerFunc — пользовательский тип обработчика, возвращающий ошибку.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// errorMiddleware оборачивает HandlerFunc, логирует ошибки и возвращает корректный JSON.
func errorMiddleware(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			log.Printf("[ERROR] %v", err)

			var httpErr *HTTPError
			if errors.As(err, &httpErr) {
				writeJSON(w, httpErr.Status, httpErr)
				return
			}

			// Неизвестная ошибка
			writeJSON(w, http.StatusInternalServerError, HTTPError{Message: "internal server error"})
		}
	}
}

// writeJSON сериализует payload в JSON и записывает его в ResponseWriter.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("[WARN] Failed to encode JSON: %v", err)
	}
}
