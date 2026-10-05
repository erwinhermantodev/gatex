package util

import (
	"testing"
	"time"
)

func TestCircuitBreakerLifecycle(t *testing.T) {
	s := &HealthStats{}
	for i := 0; i < breakerThreshold; i++ {
		if !s.ShouldAllow() {
			t.Fatal("closed circuit should allow")
		}
		s.RecordFailure()
	}
	if s.ShouldAllow() {
		t.Fatal("circuit should be open after threshold failures")
	}

	s.openedAt = time.Now().Add(-breakerCooldown)
	if !s.ShouldAllow() {
		t.Fatal("first request after cooldown should be admitted as probe")
	}
	if s.ShouldAllow() {
		t.Fatal("only one probe may be in flight while half-open")
	}

	s.RecordFailure()
	if s.GetState() != StateOpen || s.ShouldAllow() {
		t.Fatal("failed probe must reopen the circuit")
	}

	s.openedAt = time.Now().Add(-breakerCooldown)
	s.ShouldAllow()
	s.RecordSuccess()
	if s.GetState() != StateClosed || !s.ShouldAllow() {
		t.Fatal("successful probe must close the circuit")
	}
}
