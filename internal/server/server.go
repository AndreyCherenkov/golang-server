package server

import (
	"context"
	"errors"
	"golang-server/internal/api"
	"golang-server/internal/config"
	"net/http"
	"time"
)

type HTTPServer struct {
	srv *http.Server
}

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

func (s *HTTPServer) ListenAndServe() error {
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
