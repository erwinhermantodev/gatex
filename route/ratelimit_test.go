package route

import (
	"testing"
	"time"
)

func TestRateLimiterSweep(t *testing.T) {
	store := &RateLimiterStore{limiters: map[string]*limiterEntry{}}
	store.GetLimiter("1.1.1.1")
	store.GetLimiter("2.2.2.2")
	store.limiters["1.1.1.1"].lastSeen = time.Now().Add(-time.Hour)

	store.sweep(limiterIdleTTL)

	if _, ok := store.limiters["1.1.1.1"]; ok {
		t.Error("idle limiter should be evicted")
	}
	if _, ok := store.limiters["2.2.2.2"]; !ok {
		t.Error("active limiter should be kept")
	}
}
