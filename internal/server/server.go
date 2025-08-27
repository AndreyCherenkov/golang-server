package server

import (
	"errors"
	"golang-server/internal/api"
	"golang-server/internal/config"
	"net/http"
	"time"
)

// HTTPServer обёртка над http.Server с удобными методами запуска и завершения.
type HTTPServer struct {
	srv *http.Server
}

// NewHTTPServer создаёт новый HTTPServer с конфигурацией и роутером.
func NewHTTPServer(cfg config.ServerConfig, router *api.Router) *HTTPServer {
	return &HTTPServer{
		srv: &http.Server{
			Addr:         cfg.Host + ":" + cfg.Port,
			Handler:      router.Handler(),
			ReadTimeout:  time.Duration(cfg.Timeout) * time.Second,
			WriteTimeout: time.Duration(cfg.Timeout) * time.Second,
			IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
		},
	}
}

// ListenAndServe запускает HTTP сервер и возвращает ошибку,
// если сервер не был корректно остановлен.
func (s *HTTPServer) ListenAndServe() error {
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
