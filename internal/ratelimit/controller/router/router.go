package router

import (
	"fmt"
	"ivanjabrony/cloud-test/internal/logger"
	"ivanjabrony/cloud-test/internal/ratelimit/config"
	"log/slog"
	"net/http"
)

type RLRouter struct {
	cfg    *config.Config
	logger *logger.MyLogger
	router *http.ServeMux
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

	return &RLRouter{cfg, logger, r}
}

func (s *RLRouter) Run() error {
	s.logger.Info("Started serving on configured port", slog.Int("port", s.cfg.Port))
	return http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", s.cfg.Port), s.router)
}

func (s *RLRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
