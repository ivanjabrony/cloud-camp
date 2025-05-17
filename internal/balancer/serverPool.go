package balancer

import (
	"context"
	"fmt"
	"ivanjabrony/cloud-test/internal/balancer/config"
	"ivanjabrony/cloud-test/internal/logger"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	URL          *url.URL
	IsHealthy    bool
	ReverseProxy *httputil.ReverseProxy
	mu           sync.RWMutex
}

type ServerPool struct {
	servers        []*Server
	urlStrToServer map[string]*Server
	curId          uint64
}

// Global is a main pool that contains all configured servers
var Global ServerPool

// NextIndex increases curId value atomicly
func (p *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&p.curId, uint64(1)) % uint64(len(p.servers)))
}

// AddServer adds server to the ServerPool
func (p *ServerPool) AddServer(s *Server) {
	p.servers = append(p.servers, s)
	p.urlStrToServer[s.URL.String()] = s
}

// ChangeServerStatus changes a status of a backend
func (p *ServerPool) ChangeServerStatus(backendUrl *url.URL, health bool) {
	p.urlStrToServer[backendUrl.String()].SetHealth(health)
}

// GetNextServer finds a next avaliable server
func (p *ServerPool) GetNextServer() *Server {
	// loop entire backends to find out an Alive backend
	next := p.NextIndex()
	l := len(p.servers) + next // start from next and move a full cycle

	for i := next; i < l; i++ {
		idx := i % len(p.servers)       // take an index by modding
		if p.servers[idx].GetHealth() { // if we have an alive backend, use it and store if its not the original one
			if i != next {
				atomic.StoreUint64(&p.curId, uint64(idx))
			}
			return p.servers[idx]
		}
	}

	return nil
}

// NewPool creates and fully configures new server pool.
// Configuring a pool includes configuring all of inner servers, which URLs are provided via config file
func NewPool(logger *logger.MyLogger, cfg *config.Config) *ServerPool {
	logger.Info("Started configuring server pool")
	p := ServerPool{}
	p.curId = 0
	p.urlStrToServer = make(map[string]*Server, len(cfg.URLs))
	p.servers = make([]*Server, 0, len(cfg.URLs))

	for _, url := range cfg.URLs {
		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, err error) {
			logger.Error(fmt.Sprintf("[%s] %v", url.Host, err))
			logger.Error("Site unreachable", slog.Any("error", err))
			retries := GetRetryFromContext(request)

			if retries < 3 {
				select {
				case <-time.After(10 * time.Millisecond):
					ctx := context.WithValue(request.Context(), Retries, retries+1)
					proxy.ServeHTTP(writer, request.WithContext(ctx))
				}
				return
			}

			// after 3 retries, mark this backend as down
			p.ChangeServerStatus(url, false)

			// if the same request routing for few attempts with different backends, increase the count
			attempts := GetAttemptsFromContext(request)
			logger.Info("%s(%s) Attempting retry %d", request.RemoteAddr, request.URL.Path, slog.Int("Attempts", attempts))
			ctx := context.WithValue(request.Context(), Retries, attempts+1)
			LoadBalancer(logger, cfg, &p)(writer, request.WithContext(ctx))
		}

		p.AddServer(&Server{
			URL:          url,
			IsHealthy:    true,
			ReverseProxy: proxy,
		})
		logger.Info("Configured server: %s", slog.String("URL", url.String()))
	}

	return &p
}
