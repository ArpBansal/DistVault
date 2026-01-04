package loadbalancer

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Backend struct {
	URL               *url.URL
	Alive             bool
	Connections       int
	ReverseProxy      *httputil.ReverseProxy
	mux               sync.RWMutex
	totalResponseTime time.Duration
	responseCount     int64
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.Alive = alive
}

func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.Alive
}

func (b *Backend) AddConnection() {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.Connections++
}

func (b *Backend) RemoveConnection() {
	b.mux.Lock()
	defer b.mux.Unlock()
	if b.Connections > 0 {
		b.Connections--
	}
}

func (b *Backend) RecordResponseTime(duration time.Duration) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.totalResponseTime += duration
	b.responseCount++
}

func (b *Backend) GetAverageResponseTime() time.Duration {
	b.mux.RLock()
	defer b.mux.RUnlock()
	if b.responseCount == 0 {
		return 0
	}
	return b.totalResponseTime / time.Duration(b.responseCount)
}

func (b *Backend) GetConnections() int {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.Connections
}

type LoadBalancer struct {
	backends []*Backend
	mux      sync.RWMutex
}

func NewLoadBalancer(backendURLs []string) (*LoadBalancer, error) {
	backends := make([]*Backend, len(backendURLs))
	for i, rawURL := range backendURLs {
		url, err := url.Parse(rawURL)
		if err != nil {
			return nil, err
		}

		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Service is unavailable"))
		}

		backends[i] = &Backend{
			URL:          url,
			Alive:        true,
			ReverseProxy: proxy,
			Connections:  0,
		}
	}

	return &LoadBalancer{
		backends: backends,
	}, nil
}

func (lb *LoadBalancer) GetNextBackend() *Backend {
	lb.mux.RLock()
	defer lb.mux.RUnlock()

	var selected *Backend
	var minConnections int = -1
	var minResponseTime time.Duration = -1

	for _, b := range lb.backends {
		if !b.IsAlive() {
			continue
		}

		connections := b.GetConnections()
		avgResponseTime := b.GetAverageResponseTime()

		if selected == nil {
			selected = b
			minConnections = connections
			minResponseTime = avgResponseTime
			continue
		}

		if connections < minConnections ||
			(connections == minConnections && avgResponseTime < minResponseTime) {
			selected = b
			minConnections = connections
			minResponseTime = avgResponseTime
		}
	}

	if selected == nil {
		if len(lb.backends) > 0 {
			return lb.backends[0]
		}
		return nil
	}

	return selected
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lb.GetNextBackend()
	if backend == nil {
		http.Error(w, "No available backends", http.StatusServiceUnavailable)
		return
	}

	backend.AddConnection()
	defer backend.RemoveConnection()

	startTime := time.Now()
	backend.ReverseProxy.ServeHTTP(w, r)
	duration := time.Since(startTime)

	backend.RecordResponseTime(duration)
}

func (lb *LoadBalancer) HealthCheck() {
	for _, b := range lb.backends {
		status := "up"
		alive := isBackendAlive(b.URL)
		b.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("Health check on %s: %s", b.URL, status)
	}
}

func isBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func (lb *LoadBalancer) StartHealthCheck() {
	go func() {
		ticker := time.NewTicker(time.Minute)
		for {
			select {
			case <-ticker.C:
				lb.HealthCheck()
			}
		}
	}()
}
