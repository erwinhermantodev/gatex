package route

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/labstack/echo/v4"

	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/database"
	customMw "gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/route/middleware"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util"
)

// ContextRoutePathKey holds the matched route pattern (e.g. /api/v1/users/:id).
const ContextRoutePathKey = "route_path"

type compiledRoute struct {
	method  string
	pattern string
	segs    []string
	static  int
	handler echo.HandlerFunc
}

// Registry holds the DB-defined routes and swaps them atomically on Reload,
// so admin changes take effect without restarting the gateway.
type Registry struct {
	routes atomic.Pointer[[]compiledRoute]
}

func NewRegistry() *Registry {
	r := &Registry{}
	empty := []compiledRoute{}
	r.routes.Store(&empty)
	return r
}

// Reload rebuilds the route table from the DB. On error the old table is kept.
func (r *Registry) Reload() error {
	var dbRoutes []database.Route
	if err := database.GetDB().Preload("Service").Preload("ProtoMapping").Find(&dbRoutes).Error; err != nil {
		log.Printf("Route reload failed, keeping previous table: %v", err)
		return err
	}

	compiled := make([]compiledRoute, 0, len(dbRoutes))
	for _, dr := range dbRoutes {
		if dr.Path == "" || dr.Method == "" {
			continue
		}
		compiled = append(compiled, compile(dr))
	}
	// Most specific first: more static segments win, then longer patterns.
	sort.SliceStable(compiled, func(i, j int) bool {
		if compiled[i].static != compiled[j].static {
			return compiled[i].static > compiled[j].static
		}
		return len(compiled[i].segs) > len(compiled[j].segs)
	})
	r.routes.Store(&compiled)
	log.Printf("Route table loaded: %d routes", len(compiled))
	return nil
}

func compile(dr database.Route) compiledRoute {
	segs := splitPath(dr.Path)
	static := 0
	for _, s := range segs {
		if !strings.HasPrefix(s, ":") && s != "*" {
			static++
		}
	}

	var h echo.HandlerFunc = NewDynamicHandler(dr).Handle
	// Apply in reverse so the first listed middleware is outermost.
	names := middlewareNames(dr.Middleware)
	for i := len(names) - 1; i >= 0; i-- {
		mw, ok := middlewareHandler[names[i]]
		if !ok {
			// Fail closed: a typo like "jwtt" must not leave a route unprotected.
			log.Printf("Route %s %s: unknown middleware %q, route disabled", dr.Method, dr.Path, names[i])
			name := names[i]
			h = func(c echo.Context) error {
				return echo.NewHTTPError(http.StatusInternalServerError, "Route misconfigured: unknown middleware "+name)
			}
			break
		}
		h = mw(h)
	}
	h = customMw.SetContextValue(util.ContextRouterKey, dr.Tag)(h)

	return compiledRoute{
		method:  strings.ToUpper(dr.Method),
		pattern: dr.Path,
		segs:    segs,
		static:  static,
		handler: h,
	}
}

func splitPath(p string) []string {
	return strings.Split(strings.Trim(p, "/"), "/")
}

func match(segs, parts []string) (names, values []string, ok bool) {
	for i, s := range segs {
		if s == "*" && i == len(segs)-1 {
			return append(names, "*"), append(values, strings.Join(parts[min(i, len(parts)):], "/")), true
		}
		if i >= len(parts) {
			return nil, nil, false
		}
		if strings.HasPrefix(s, ":") {
			if parts[i] == "" {
				return nil, nil, false
			}
			names = append(names, s[1:])
			values = append(values, parts[i])
		} else if s != parts[i] {
			return nil, nil, false
		}
	}
	return names, values, len(segs) == len(parts)
}

// Handle is the catch-all handler: it resolves the request against the
// current route table.
func (r *Registry) Handle(c echo.Context) error {
	parts := splitPath(c.Request().URL.Path)
	method := c.Request().Method
	pathMatched := false

	for _, cr := range *r.routes.Load() {
		names, values, ok := match(cr.segs, parts)
		if !ok {
			continue
		}
		if cr.method != method {
			pathMatched = true
			continue
		}
		c.SetParamNames(names...)
		c.SetParamValues(values...)
		c.Set(ContextRoutePathKey, cr.pattern)
		return cr.handler(c)
	}

	if pathMatched {
		return echo.NewHTTPError(http.StatusMethodNotAllowed)
	}
	return echo.ErrNotFound
}
