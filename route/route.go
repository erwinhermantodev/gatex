package route

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/config"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain"
	adminHandler "gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain/admin/handler"
	customMw "gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/route/middleware"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util"
)

// Route for mapping from json file
type Route struct {
	Path       string   `json:"path"`
	Method     string   `json:"method"`
	Module     string   `json:"module"`
	Tag        string   `json:"tag"`
	Endpoint   string   `json:"endpoint_filter"`
	Middleware []string `json:"middleware"`
}

// Redundant definition removed, moved to domain

// Init gateway router
func Init() *echo.Echo {
	cfg := config.Load()
	registry := NewRegistry()
	if err := registry.Reload(); err != nil {
		panic(err)
	}

	e := echo.New()
	if !cfg.TrustProxyHeaders {
		// Use the TCP peer address; X-Forwarded-For is client-controlled.
		e.IPExtractor = echo.ExtractIPDirect()
	}
	e.Validator = &domain.CustomValidator{Validator: validator.New()}

	store := NewRateLimiterStore()
	// RequestID must run first so every later middleware sees the ID.
	e.Use(middleware.RequestID())
	e.Use(CacheControlMiddleware)
	e.Use(customMw.MetricsMiddleware)
	e.Use(customMw.TrafficLogger())
	e.Use(rateLimiterMiddleware(store))
	// Set Bundle MiddleWare
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  cfg.CORSAllowOrigins,
		AllowHeaders:  []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderContentLength, echo.HeaderAcceptEncoding, echo.HeaderAccessControlAllowOrigin, echo.HeaderAccessControlAllowHeaders, echo.HeaderContentDisposition, "X-Request-Id", "device-id", "X-Summary", "X-Account-Number", "X-Business-Name", "client-secret", "X-CSRF-Token", "x-api-key", "Cache-Control"},
		ExposeHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderContentLength, echo.HeaderAcceptEncoding, echo.HeaderAccessControlAllowOrigin, echo.HeaderAccessControlAllowHeaders, echo.HeaderContentDisposition, "X-Request-Id", "device-id", "X-Summary", "X-Account-Number", "X-Business-Name", "client-secret", "X-CSRF-Token", "x-api-key", "Cache-Control"},
		AllowMethods:  []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))

	e.HTTPErrorHandler = util.CustomHTTPErrorHandler

	// Register Admin API (token protected)
	admin := adminHandler.NewAdminHandler(func() { _ = registry.Reload() })
	a := e.Group("/admin", customMw.AdminAuth(cfg.AdminAPIToken))

	// Services
	a.GET("/services", admin.GetServices)
	a.POST("/services", admin.CreateService)
	a.PUT("/services/:id", admin.UpdateService)
	a.DELETE("/services/:id", admin.DeleteService)

	// Routes
	a.GET("/routes", admin.GetRoutes)
	a.POST("/routes", admin.CreateRoute)
	a.PUT("/routes/:id", admin.UpdateRoute)
	a.DELETE("/routes/:id", admin.DeleteRoute)

	// Proto Mappings
	a.GET("/proto-mappings", admin.GetProtoMappings)
	a.POST("/proto-mappings", admin.CreateProtoMapping)
	a.PUT("/proto-mappings/:id", admin.UpdateProtoMapping)
	a.DELETE("/proto-mappings/:id", admin.DeleteProtoMapping)
	a.GET("/metrics", admin.GetMetrics)
	a.GET("/logs", admin.GetActivityLogs)
	a.GET("/request-logs", admin.GetRequestLogs)
	a.GET("/traces/:id", admin.GetTraceLogs)
	a.GET("/server-logs", admin.GetServerLogs)

	// Serve Dashboard
	e.Static("/dashboard", "dashboard/dist")
	e.File("/dashboard", "dashboard/dist/index.html")

	// Everything else is resolved against the DB-backed route table.
	e.Any("/*", registry.Handle)

	return e
}

// CacheControlMiddleware sets cache control headers
func CacheControlMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
		return next(c)
	}
}

const (
	rateLimitPerSecond = 10
	rateLimitBurst     = 5
	limiterIdleTTL     = 10 * time.Minute
	limiterSweepEvery  = time.Minute
)

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiterStore keeps one rate limiter per IP and evicts idle ones.
type RateLimiterStore struct {
	limiters map[string]*limiterEntry
	mutex    sync.Mutex
}

// NewRateLimiterStore creates a store and starts its background sweeper.
func NewRateLimiterStore() *RateLimiterStore {
	store := &RateLimiterStore{limiters: make(map[string]*limiterEntry)}
	go func() {
		for range time.Tick(limiterSweepEvery) {
			store.sweep(limiterIdleTTL)
		}
	}()
	return store
}

// GetLimiter retrieves the rate limiter for a specific IP, creating one if necessary
func (store *RateLimiterStore) GetLimiter(ip string) *rate.Limiter {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	entry, exists := store.limiters[ip]
	if !exists {
		entry = &limiterEntry{limiter: rate.NewLimiter(rate.Limit(rateLimitPerSecond), rateLimitBurst)}
		store.limiters[ip] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter
}

// sweep drops limiters not used for longer than ttl.
func (store *RateLimiterStore) sweep(ttl time.Duration) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	for ip, entry := range store.limiters {
		if time.Since(entry.lastSeen) > ttl {
			delete(store.limiters, ip)
		}
	}
}

// rateLimiterMiddleware creates the rate limiting middleware
func rateLimiterMiddleware(store *RateLimiterStore) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip rate limiting for admin and dashboard
			path := c.Request().URL.Path
			if util.IsAdminPath(path) || strings.HasPrefix(path, "/dashboard") {
				return next(c)
			}

			ip := c.RealIP()
			limiter := store.GetLimiter(ip)

			if !limiter.Allow() {
				return c.String(http.StatusTooManyRequests, "Too Many Requests")
			}

			return next(c)
		}
	}
}
