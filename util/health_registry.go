package util

import (
	"sync"
	"time"
)

type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

type HealthStats struct {
	TotalRequests       int
	FailedRequests      int
	LastFailure         time.Time
	State               CircuitState
	ConsecutiveFailures int
	mu                  sync.RWMutex
}

var (
	Registry   = make(map[uint]*HealthStats)
	registryMu sync.RWMutex
)

func GetHealthStats(serviceID uint) *HealthStats {
	registryMu.RLock()
	stats, ok := Registry[serviceID]
	registryMu.RUnlock()

	if ok {
		return stats
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	// Double check
	if stats, ok = Registry[serviceID]; ok {
		return stats
	}

	stats = &HealthStats{
		State: StateClosed,
	}
	Registry[serviceID] = stats
	return stats
}

func (s *HealthStats) RecordSuccess() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalRequests++
	s.ConsecutiveFailures = 0
	if s.State == StateHalfOpen {
		s.State = StateClosed
	}
}

func (s *HealthStats) RecordFailure() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalRequests++
	s.FailedRequests++
	s.ConsecutiveFailures++
	s.LastFailure = time.Now()

	// Simple Circuit Breaking Logic: 5 consecutive failures trips it
	if s.ConsecutiveFailures >= 5 && s.State == StateClosed {
		s.State = StateOpen
		// In a real system, we'd use a timer to move to HalfOpen
		go func() {
			time.Sleep(30 * time.Second)
			s.mu.Lock()
			if s.State == StateOpen {
				s.State = StateHalfOpen
			}
			s.mu.Unlock()
		}()
	}
}

func (s *HealthStats) ShouldAllow() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.State != StateOpen
}

func (s *HealthStats) GetHealthScore() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.TotalRequests == 0 {
		return 100
	}
	return ((s.TotalRequests - s.FailedRequests) * 100) / s.TotalRequests
}
