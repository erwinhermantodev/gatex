package route

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/database"
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
	routes := loadRoutesFromDB()

	e := echo.New()
	e.Validator = &domain.CustomValidator{Validator: validator.New()}

	store := NewRateLimiterStore()
	e.Use(CacheControlMiddleware)
	e.Use(customMw.MetricsMiddleware)
	e.Use(customMw.TrafficLogger())
	e.Use(rateLimiterMiddleware(store))
	// Set Bundle MiddleWare
	e.Use(middleware.RequestID())
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowHeaders:  []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderContentLength, echo.HeaderAcceptEncoding, echo.HeaderAccessControlAllowOrigin, echo.HeaderAccessControlAllowHeaders, echo.HeaderContentDisposition, "X-Request-Id", "device-id", "X-Summary", "X-Account-Number", "X-Business-Name", "client-secret", "X-CSRF-Token", "x-api-key", "Cache-Control", "no-store, no-cache, must-revalidate, private"},
		ExposeHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderContentLength, echo.HeaderAcceptEncoding, echo.HeaderAccessControlAllowOrigin, echo.HeaderAccessControlAllowHeaders, echo.HeaderContentDisposition, "X-Request-Id", "device-id", "X-Summary", "X-Account-Number", "X-Business-Name", "client-secret", "X-CSRF-Token", "x-api-key", "Cache-Control", "no-store, no-cache, must-revalidate, private"},
		AllowMethods:  []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))

	e.HTTPErrorHandler = util.CustomHTTPErrorHandler

	for _, route := range routes {
		h := NewDynamicHandler(route.Endpoint)
		e.Add(route.Method, route.Path, h.Handle, chainMiddleware(route)...)
	}

	// Register Admin API
	admin := adminHandler.NewAdminHandler()
	a := e.Group("/admin")

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

	return e
}

func loadRoutesFromDB() []Route {
	db := database.GetDB()
	var dbRoutes []database.Route
	if err := db.Find(&dbRoutes).Error; err != nil {
		log.Printf("Error loading routes from DB: %v", err)
		return nil
	}

	var routes []Route
	for _, dr := range dbRoutes {
		var mw []string
		_ = json.Unmarshal([]byte(dr.Middleware), &mw)
		routes = append(routes, Route{
			Path:       dr.Path,
			Method:     dr.Method,
			Tag:        dr.Tag,
			Endpoint:   dr.EndpointFilter,
			Middleware: mw,
		})
	}

	return routes
}

func chainMiddleware(route Route) []echo.MiddlewareFunc {
	var mwHandlers []echo.MiddlewareFunc
	// init mw for router ,attach router properties
	mwHandlers = append(mwHandlers, customMw.SetContextValue(util.ContextRouterKey, route.Tag))
	for _, v := range route.Middleware {
		mwHandlers = append(mwHandlers, middlewareHandler[v])
	}
	return mwHandlers
}

// CacheControlMiddleware sets cache control headers
func CacheControlMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
		return next(c)
	}
}

// RateLimiterStore to store rate limiters per IP
type RateLimiterStore struct {
	limiters map[string]*rate.Limiter
	mutex    sync.Mutex
}

// NewRateLimiterStore creates a new RateLimiterStore
func NewRateLimiterStore() *RateLimiterStore {
	return &RateLimiterStore{
		limiters: make(map[string]*rate.Limiter),
	}
}

// GetLimiter retrieves the rate limiter for a specific IP, creating one if necessary
func (store *RateLimiterStore) GetLimiter(ip string) *rate.Limiter {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	limiter, exists := store.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(10), 5) // 2 requests per second with a burst of 5
		store.limiters[ip] = limiter
	}

	return limiter
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
