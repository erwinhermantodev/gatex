package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// TimeoutMiddleware sets a context timeout for the request
func TimeoutMiddleware(timeout time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx, cancel := context.WithTimeout(c.Request().Context(), timeout)
			defer cancel()

			c.SetRequest(c.Request().WithContext(ctx))

			done := make(chan error, 1)
			go func() {
				done <- next(c)
			}()

			select {
			case err := <-done:
				return err
			case <-ctx.Done():
				return echo.NewHTTPError(http.StatusGatewayTimeout, "Gateway Timeout")
			}
		}
	}
}

// RetryMiddleware retries the request if it fails with a 5xx error
func RetryMiddleware(maxRetries int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var err error
			for i := 0; i < maxRetries; i++ {
				err = next(c)
				if err == nil {
					return nil
				}

				// Only retry on server errors or timeouts
				he, ok := err.(*echo.HTTPError)
				if ok && he.Code < 500 && he.Code != http.StatusRequestTimeout {
					return err
				}

				// Wait before retry
				time.Sleep(time.Duration(i*100) * time.Millisecond)
			}
			return err
		}
	}
}

// Simple Circuit Breaker
type circuitBreaker struct {
	failures     int
	threshold    int
	lastFailure  time.Time
	resetTimeout time.Duration
	mu           sync.RWMutex
}

var breakers = make(map[string]*circuitBreaker)
var breakersMu sync.RWMutex

func getBreaker(service string) *circuitBreaker {
	breakersMu.Lock()
	defer breakersMu.Unlock()

	if _, ok := breakers[service]; !ok {
		breakers[service] = &circuitBreaker{
			threshold:    5,
			resetTimeout: 30 * time.Second,
		}
	}
	return breakers[service]
}

func CircuitBreakerMiddleware(service string) echo.MiddlewareFunc {
	cb := getBreaker(service)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cb.mu.RLock()
			if cb.failures >= cb.threshold && time.Since(cb.lastFailure) < cb.resetTimeout {
				cb.mu.RUnlock()
				return echo.NewHTTPError(http.StatusServiceUnavailable, "Circuit breaker open for service: "+service)
			}
			cb.mu.RUnlock()

			err := next(c)

			cb.mu.Lock()
			defer cb.mu.Unlock()
			if err != nil {
				he, ok := err.(*echo.HTTPError)
				if ok && he.Code >= 500 {
					cb.failures++
					cb.lastFailure = time.Now()
				}
			} else {
				cb.failures = 0
			}

			return err
		}
	}
}
