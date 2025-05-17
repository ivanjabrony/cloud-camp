package balancer

import (
	"context"
	"fmt"
	"ivanjabrony/cloud-test/internal/balancer/config"
	"ivanjabrony/cloud-test/internal/logger"
	"log/slog"
	"net"
	"time"
)

const (
	Attempts = "attempts"
	Retries  = "retries"
)

func (s *Server) GetHealth() bool {
	s.mu.RLock()
	health := s.IsHealthy
	s.mu.RUnlock()

	return health
}

func (s *Server) SetHealth(health bool) {
	s.mu.Lock()
	s.IsHealthy = health
	s.mu.Unlock()
}

// HealthCheck checks and updates Service availability
func (s *Server) HealthCheck(logger *logger.MyLogger, cfg *config.Config) bool {
	timeout := cfg.HealthServerTimeout
	conn, err := net.DialTimeout("tcp", s.URL.Host, timeout)
	if err != nil {
		logger.Error("Site unreachable", slog.Any("error", err))
		return false
	}
	defer conn.Close()

	return true
}

// HealthCheck checks and updates ServerPool availability
func (p *ServerPool) HealthCheck(logger *logger.MyLogger, cfg *config.Config) {
	for _, s := range p.servers {
		status := "up"
		isHealthy := s.HealthCheck(logger, cfg)
		s.SetHealth(isHealthy)
		if !isHealthy {
			status = "down"
		}

		logger.Info(fmt.Sprintf("%s [%s]", s.URL.String(), status))
	}
}

// HealthCheckRoutine is a goroutine that checks health of every server and updates it based on a Healthcheck func result
func HealthCheckRoutine(ctx context.Context, logger *logger.MyLogger, cfg *config.Config, pool *ServerPool) { // logger
	t := time.NewTicker(cfg.HealthPoolTimeout)
	defer t.Stop()
	logger.Info("Ticker timout:", slog.Float64("seconds", cfg.HealthPoolTimeout.Seconds()))
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			logger.Info("Starting health check...")
			pool.HealthCheck(logger, cfg)
			logger.Info("Health check completed")
		}
	}
}
