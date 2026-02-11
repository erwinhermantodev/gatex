package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/database"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util/logbuffer"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util/metrics"
)

type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

// --- Service Handlers ---

func (h *AdminHandler) GetServices(c echo.Context) error {
	var services []database.Service
	db := database.GetDB()
	if err := db.Find(&services).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, services)
}

func (h *AdminHandler) CreateService(c echo.Context) error {
	service := new(database.Service)
	if err := c.Bind(service); err != nil {
		return err
	}
	db := database.GetDB()
	if err := db.Create(service).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	util.LogCreate("Service", "admin", service.Name)
	return c.JSON(http.StatusCreated, service)
}

func (h *AdminHandler) UpdateService(c echo.Context) error {
	id := c.Param("id")
	var service database.Service
	db := database.GetDB()
	if err := db.First(&service, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Service not found")
	}
	if err := c.Bind(&service); err != nil {
		return err
	}
	db.Save(&service)
	util.LogUpdate("Service", "admin", service.Name)
	return c.JSON(http.StatusOK, service)
}

func (h *AdminHandler) DeleteService(c echo.Context) error {
	id := c.Param("id")
	db := database.GetDB()
	if err := db.Delete(&database.Service{}, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	util.LogDelete("Service", "admin", "ID: "+id)
	return c.NoContent(http.StatusNoContent)
}

// --- Route Handlers ---

func (h *AdminHandler) GetRoutes(c echo.Context) error {
	var routes []database.Route
	db := database.GetDB()
	if err := db.Preload("Service").Find(&routes).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, routes)
}

func (h *AdminHandler) CreateRoute(c echo.Context) error {
	route := new(database.Route)
	if err := c.Bind(route); err != nil {
		return err
	}
	db := database.GetDB()
	if err := db.Create(route).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	util.LogCreate("Route", "admin", route.Path)
	return c.JSON(http.StatusCreated, route)
}

func (h *AdminHandler) UpdateRoute(c echo.Context) error {
	id := c.Param("id")
	var route database.Route
	db := database.GetDB()
	if err := db.First(&route, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Route not found")
	}
	if err := c.Bind(&route); err != nil {
		return err
	}
	db.Save(&route)
	util.LogUpdate("Route", "admin", route.Path)
	return c.JSON(http.StatusOK, route)
}

func (h *AdminHandler) DeleteRoute(c echo.Context) error {
	id := c.Param("id")
	db := database.GetDB()
	if err := db.Delete(&database.Route{}, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	util.LogDelete("Route", "admin", "ID: "+id)
	return c.NoContent(http.StatusNoContent)
}

// --- Proto Mapping Handlers ---

func (h *AdminHandler) GetProtoMappings(c echo.Context) error {
	var mappings []database.ProtoMapping
	db := database.GetDB()
	if err := db.Preload("Service").Find(&mappings).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, mappings)
}

func (h *AdminHandler) CreateProtoMapping(c echo.Context) error {
	mapping := new(database.ProtoMapping)
	if err := c.Bind(mapping); err != nil {
		return err
	}
	db := database.GetDB()
	if err := db.Create(mapping).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, mapping)
}

func (h *AdminHandler) UpdateProtoMapping(c echo.Context) error {
	id := c.Param("id")
	var mapping database.ProtoMapping
	db := database.GetDB()
	if err := db.First(&mapping, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "ProtoMapping not found")
	}
	if err := c.Bind(&mapping); err != nil {
		return err
	}
	db.Save(&mapping)
	return c.JSON(http.StatusOK, mapping)
}

func (h *AdminHandler) DeleteProtoMapping(c echo.Context) error {
	id := c.Param("id")
	db := database.GetDB()
	if err := db.Delete(&database.ProtoMapping{}, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	util.LogDelete("ProtoMapping", "admin", "ID: "+id)
	return c.NoContent(http.StatusNoContent)
}

func (h *AdminHandler) GetActivityLogs(c echo.Context) error {
	var logs []database.ActivityLog
	db := database.GetDB()
	if err := db.Order("id desc").Limit(50).Find(&logs).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, logs)
}
func (h *AdminHandler) GetRequestLogs(c echo.Context) error {
	var logs []database.RequestLog
	db := database.GetDB()
	if err := db.Order("id desc").Limit(100).Find(&logs).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, logs)
}

func (h *AdminHandler) GetTraceLogs(c echo.Context) error {
	requestID := c.Param("id")
	var logs []database.TraceLog
	db := database.GetDB()
	if err := db.Where("request_id = ?", requestID).Order("id asc").Find(&logs).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, logs)
}

func (h *AdminHandler) GetServerLogs(c echo.Context) error {
	return c.JSON(http.StatusOK, logbuffer.DefaultBuffer.GetEntries())
}

func (h *AdminHandler) GetMetrics(c echo.Context) error {
	// Populate health info from registry before returning
	db := database.GetDB()
	var services []database.Service
	db.Find(&services)

	for _, s := range services {
		stats := util.GetHealthStats(s.ID)
		m := metrics.DefaultRegistry.GetServiceMetrics(s.Name)
		m.HealthScore = stats.GetHealthScore()

		status := "CLOSED"
		switch stats.State {
		case util.StateOpen:
			status = "OPEN"
		case util.StateHalfOpen:
			status = "HALF-OPEN"
		}
		m.CircuitStatus = status
	}

	return c.JSON(http.StatusOK, metrics.DefaultRegistry)
}
