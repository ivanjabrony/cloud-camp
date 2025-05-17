package router

import (
	"context"
	"fmt"
	"ivanjabrony/cloud-test/internal/logger"
	"ivanjabrony/cloud-test/internal/ratelimit/config"
	"log/slog"
	"net/http"
)

type RLRouter struct {
	cfg    *config.Config
	logger *logger.MyLogger
	server *http.Server
}

type ConfigHandler interface {
	UpdateConfiguration(w http.ResponseWriter, r *http.Request)
}

type RateLimitHandler interface {
	RateLimit(w http.ResponseWriter, r *http.Request)
}

func NewRouter(cfg *config.Config, logger *logger.MyLogger, configHandler ConfigHandler, rlHandler RateLimitHandler) *RLRouter {
	r := http.NewServeMux()

	r.HandleFunc("POST /config", configHandler.UpdateConfiguration)
	r.HandleFunc("/", rlHandler.RateLimit)

	server := http.Server{Addr: fmt.Sprintf("0.0.0.0:%d", cfg.Port), Handler: r}
	return &RLRouter{cfg, logger, &server}
}

func (s *RLRouter) Run() error {
	s.logger.Info("Started serving on configured port", slog.Int("port", s.cfg.Port))
	return s.server.ListenAndServe()
}

func (s *RLRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.server.Handler.ServeHTTP(w, r)
}

func (s *RLRouter) Stop(ctx context.Context) error {
	s.logger.Info("Started shutting down router")
	return s.server.Shutdown(ctx)
}
