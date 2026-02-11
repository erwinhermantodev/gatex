package metrics

import (
	"sync"
	"time"
)

type MetricPoint struct {
	Count     int64
	Latency   time.Duration
	Errors    int64
	Timestamp time.Time
}

type ServiceMetrics struct {
	TotalRequests int64                `json:"total_requests"`
	TotalErrors   int64                `json:"total_errors"`
	AvgLatencyMS  float64              `json:"avg_latency_ms"`
	LastStatus    int                  `json:"last_status"`
	StatusCounts  map[int]int64        `json:"status_counts"`
	PathMetrics   map[string]*PathInfo `json:"path_metrics"`
	HealthScore   int                  `json:"health_score"`
	CircuitStatus string               `json:"circuit_status"`
	mu            sync.RWMutex
}

type PathInfo struct {
	Count        int64   `json:"count"`
	AvgLatencyMS float64 `json:"avg_latency_ms"`
}

type Registry struct {
	Services  map[string]*ServiceMetrics `json:"services"`
	StartTime time.Time                  `json:"start_time"`
	mu        sync.RWMutex
}

var DefaultRegistry = &Registry{
	Services:  make(map[string]*ServiceMetrics),
	StartTime: time.Now(),
}

func (r *Registry) GetServiceMetrics(serviceName string) *ServiceMetrics {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.Services[serviceName]; !ok {
		r.Services[serviceName] = &ServiceMetrics{
			StatusCounts: make(map[int]int64),
			PathMetrics:  make(map[string]*PathInfo),
		}
	}
	return r.Services[serviceName]
}

func (sm *ServiceMetrics) Record(path string, status int, duration time.Duration) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.TotalRequests++
	if status >= 400 {
		sm.TotalErrors++
	}
	sm.LastStatus = status
	sm.StatusCounts[status]++

	// Update Service Latency
	ms := float64(duration.Microseconds()) / 1000.0
	if sm.AvgLatencyMS == 0 {
		sm.AvgLatencyMS = ms
	} else {
		// Moving average (simplified)
		sm.AvgLatencyMS = (sm.AvgLatencyMS * 0.9) + (ms * 0.1)
	}

	// Update Path Metrics
	if _, ok := sm.PathMetrics[path]; !ok {
		sm.PathMetrics[path] = &PathInfo{}
	}
	pi := sm.PathMetrics[path]
	pi.Count++
	if pi.AvgLatencyMS == 0 {
		pi.AvgLatencyMS = ms
	} else {
		pi.AvgLatencyMS = (pi.AvgLatencyMS * 0.9) + (ms * 0.1)
	}
}

func Record(service, path string, status int, duration time.Duration) {
	DefaultRegistry.GetServiceMetrics(service).Record(path, status, duration)
}
