package handler

import (
	"context"
	"errors"
	"ivanjabrony/cloud-test/internal/logger"
	"ivanjabrony/cloud-test/internal/ratelimit/config"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type RateLimiter interface {
	Allow(ctx context.Context, ip string) bool
}

type RateLimitProxy struct {
	targetURL *url.URL
	proxy     *httputil.ReverseProxy
}

type RateLimitHandler struct {
	cfg         *config.Config
	logger      *logger.MyLogger
	proxy       *RateLimitProxy
	rateLimiter RateLimiter
}

func NewRateLimitHandler(cfg *config.Config, logger *logger.MyLogger, ratelimiter RateLimiter) (*RateLimitHandler, error) {
	if cfg == nil || logger == nil || ratelimiter == nil {
		return nil, errors.New("nil values in handler constructor")
	}

	return &RateLimitHandler{cfg, logger, newRateLimitProxy(cfg.TargetURL), ratelimiter}, nil
}

func newRateLimitProxy(url *url.URL) *RateLimitProxy {
	return &RateLimitProxy{
		targetURL: url,
		proxy:     httputil.NewSingleHostReverseProxy(url),
	}
}

func (rl *RateLimitHandler) RateLimit(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientID(r)

	// Проверка rate limit
	if !rl.rateLimiter.Allow(r.Context(), clientIP) {
		rl.logger.Warn("Rate limit exceeded", slog.String("client", clientIP))
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	rl.logger.Debug("Proxying request",
		slog.String("client", clientIP),
		slog.String("path", r.URL.Path))

	// Перенаправление запроса в целевой сервис
	rl.proxy.proxy.ServeHTTP(w, r)
}

func getClientID(r *http.Request) string {
	if id := r.Header.Get("X-Client-ID"); id != "" {
		return id
	}
	return r.RemoteAddr
}
