package route

import (
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/domain"
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
	routes := loadRoutes("./route/gate/")

	e := echo.New()
	e.Validator = &domain.CustomValidator{Validator: validator.New()}

	store := NewRateLimiterStore()
	e.Use(CacheControlMiddleware)
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
		e.Add(route.Method, route.Path, endpoint[route.Endpoint].Handle, chainMiddleware(route)...)
	}

	return e
}

func loadRoutes(filePath string) []Route {
	var routes []Route
	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		log.Fatalf("Failed to load file: %v", err)
	}
	for _, file := range files {
		byteFile, err := ioutil.ReadFile(filePath + "/" + file.Name())
		if err != nil {
			log.Fatalf("Failed to load file: %v", err)
		}
		var tmp []Route
		if err := util.Json.Unmarshal(byteFile, &tmp); err != nil {
			log.Fatalf("Failed to marshal file: %v", err)
		}
		routes = append(routes, tmp...)
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
			ip := c.RealIP()
			limiter := store.GetLimiter(ip)

			if !limiter.Allow() {
				return c.String(http.StatusTooManyRequests, "Too Many Requests")
			}

			return next(c)
		}
	}
}
