package handler

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/database"
	customMw "gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/route/middleware"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util/netguard"
	"gorm.io/gorm"
)

func badRequest(format string, args ...interface{}) error {
	return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(format, args...))
}

var allowedMethods = map[string]bool{"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true, "HEAD": true}

// validateService checks and normalises a service before it is saved.
func validateService(s *database.Service) error {
	s.Name = strings.TrimSpace(s.Name)
	if s.Name == "" {
		return badRequest("Name is required")
	}
	s.Protocol = strings.ToLower(strings.TrimSpace(s.Protocol))
	if s.Protocol == "" {
		s.Protocol = "rest"
	}

	switch s.Protocol {
	case "rest":
		u, err := url.Parse(strings.TrimSpace(s.BaseURL))
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
			return badRequest("BaseURL must be an absolute http(s) URL")
		}
		if u.User != nil {
			return badRequest("BaseURL must not contain credentials")
		}
		if err := netguard.CheckHost(u.Hostname()); err != nil {
			return badRequest("BaseURL rejected: %v", err)
		}
		s.BaseURL = strings.TrimRight(u.String(), "/")
	case "grpc":
		host, port, err := net.SplitHostPort(strings.TrimSpace(s.GRPCAddr))
		if err != nil || host == "" || port == "" {
			return badRequest("GRPCAddr must be host:port")
		}
		if err := netguard.CheckHost(host); err != nil {
			return badRequest("GRPCAddr rejected: %v", err)
		}
		s.GRPCAddr = net.JoinHostPort(host, port)
	default:
		return badRequest("Protocol must be \"rest\" or \"grpc\"")
	}
	return nil
}

// validateRoutePath rejects reserved prefixes and malformed patterns.
func validateRoutePath(path string) error {
	if !strings.HasPrefix(path, "/") || strings.Contains(path, "//") || strings.ContainsAny(path, " ?#") {
		return badRequest("Path must start with / and contain no spaces, '//', '?' or '#'")
	}
	for _, reserved := range []string{"/admin", "/dashboard"} {
		if path == reserved || strings.HasPrefix(path, reserved+"/") {
			return badRequest("Path %s is reserved", reserved)
		}
	}
	segs := strings.Split(strings.Trim(path, "/"), "/")
	for i, seg := range segs {
		if seg == ":" {
			return badRequest("Path parameter needs a name")
		}
		if strings.Contains(seg, "*") && (seg != "*" || i != len(segs)-1) {
			return badRequest("'*' is only allowed as the last whole segment")
		}
	}
	return nil
}

// validateRoute checks a route and the records it references.
func validateRoute(db *gorm.DB, r *database.Route) error {
	r.Method = strings.ToUpper(strings.TrimSpace(r.Method))
	if !allowedMethods[r.Method] {
		return badRequest("Method must be one of GET, POST, PUT, PATCH, DELETE, HEAD")
	}
	r.Path = strings.TrimSpace(r.Path)
	if err := validateRoutePath(r.Path); err != nil {
		return err
	}

	var svc database.Service
	if err := db.First(&svc, r.ServiceID).Error; err != nil {
		return badRequest("ServiceID %d does not exist", r.ServiceID)
	}

	if r.ProtoMappingID != nil {
		var m database.ProtoMapping
		if err := db.First(&m, *r.ProtoMappingID).Error; err != nil {
			return badRequest("ProtoMappingID %d does not exist", *r.ProtoMappingID)
		}
		if m.ServiceID != r.ServiceID {
			return badRequest("ProtoMapping belongs to a different service")
		}
	}

	if strings.TrimSpace(r.Middleware) == "" {
		r.Middleware = "[]"
	}
	var names []string
	if err := json.Unmarshal([]byte(r.Middleware), &names); err != nil {
		return badRequest("Middleware must be a JSON array of names")
	}
	known := map[string]bool{}
	for _, n := range customMw.KnownNames {
		known[n] = true
	}
	for _, n := range names {
		if !known[n] {
			return badRequest("Unknown middleware %q (known: %s)", n, strings.Join(customMw.KnownNames, ", "))
		}
	}
	return nil
}

func validateProtoMapping(db *gorm.DB, m *database.ProtoMapping) error {
	var svc database.Service
	if err := db.First(&svc, m.ServiceID).Error; err != nil {
		return badRequest("ServiceID %d does not exist", m.ServiceID)
	}
	if svc.Protocol != "grpc" {
		return badRequest("Service %q is not a grpc service", svc.Name)
	}
	for field, v := range map[string]string{"RPCMethod": m.RPCMethod, "ServiceName": m.ServiceName, "ProtoPackage": m.ProtoPackage} {
		if strings.TrimSpace(v) == "" {
			return badRequest("%s is required", field)
		}
	}
	return nil
}
