package balancer

import (
	"ivanjabrony/cloud-test/internal/balancer/config"
	"ivanjabrony/cloud-test/internal/logger"
	"net/http"
)

func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value("Attempts").(int); ok {
		return attempts
	}
	return 1
}

func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value("Retry").(int); ok {
		return retry
	}
	return 0
}

// LoadBalancer return a func that balances request between available services and checks availability
func LoadBalancer(logger *logger.MyLogger, cfg *config.Config, pool *ServerPool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		attempts := GetAttemptsFromContext(r)
		if attempts > cfg.MaxAttempts {
			logger.Warn("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
			http.Error(w, "Service not available", http.StatusServiceUnavailable)
			return
		}

		peer := pool.GetNextServer()
		if peer != nil {
			peer.ReverseProxy.ServeHTTP(w, r)
			return
		}
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
	}
}
