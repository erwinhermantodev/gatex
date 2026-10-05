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

const (
	breakerThreshold = 5
	breakerCooldown  = 30 * time.Second
)

type HealthStats struct {
	TotalRequests       int
	FailedRequests      int
	LastFailure         time.Time
	State               CircuitState
	ConsecutiveFailures int
	openedAt            time.Time
	probeStarted        time.Time // zero when no half-open probe is in flight
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
		s.probeStarted = time.Time{}
	}
}

func (s *HealthStats) RecordFailure() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalRequests++
	s.FailedRequests++
	s.ConsecutiveFailures++
	s.LastFailure = time.Now()

	switch s.State {
	case StateClosed:
		if s.ConsecutiveFailures >= breakerThreshold {
			s.trip()
		}
	case StateHalfOpen:
		// The probe failed: reopen for another cooldown.
		s.trip()
	}
}

// trip opens the circuit. Callers must hold s.mu.
func (s *HealthStats) trip() {
	s.State = StateOpen
	s.openedAt = time.Now()
	s.probeStarted = time.Time{}
}

// ShouldAllow reports whether a request may go upstream. After the cooldown
// the circuit moves to half-open and admits a single probe request; its
// outcome (RecordSuccess/RecordFailure) closes or reopens the circuit.
func (s *HealthStats) ShouldAllow() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.State == StateOpen && time.Since(s.openedAt) >= breakerCooldown {
		s.State = StateHalfOpen
		s.probeStarted = time.Time{}
	}

	switch s.State {
	case StateOpen:
		return false
	case StateHalfOpen:
		// Admit one probe; if it never reported back, allow another after a cooldown.
		if s.probeStarted.IsZero() || time.Since(s.probeStarted) >= breakerCooldown {
			s.probeStarted = time.Now()
			return true
		}
		return false
	}
	return true
}

// GetState returns the current circuit state.
func (s *HealthStats) GetState() CircuitState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.State
}

func (s *HealthStats) GetHealthScore() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.TotalRequests == 0 {
		return 100
	}
	return ((s.TotalRequests - s.FailedRequests) * 100) / s.TotalRequests
}
