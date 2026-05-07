package handler

import (
	"net/http"
	"strconv"
	"time"

	"aether-node/internal/domain/alarm"
	"aether-node/pkg/response"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

type AlarmHandler struct {
	svc alarm.AlarmService
}

func NewAlarmHandler(svc alarm.AlarmService) *AlarmHandler {
	return &AlarmHandler{svc: svc}
}

func (h *AlarmHandler) GetActiveAlarms(c echo.Context) error {
	ctx := c.Request().Context()
	deviceGUID := c.Param("device_guid")

	results, err := h.svc.ListActiveAlarms(ctx, deviceGUID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, results, "Active alarms retrieved successfully")
}

func (h *AlarmHandler) GetAlarmHistory(c echo.Context) error {
	ctx := c.Request().Context()
	
	deviceGUID := c.QueryParam("device_guid")
	locationGUID := c.QueryParam("location_guid")
	status := c.QueryParam("status")
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	page, _ := strconv.Atoi(c.QueryParam("page"))

	params := alarm.ListParams{
		Limit: limit,
		Page:  page,
	}
	if deviceGUID != "" {
		params.DeviceGUID = &deviceGUID
	}
	if locationGUID != "" {
		params.LocationGUID = &locationGUID
	}
	if status != "" {
		params.Status = &status
	}

	result, err := h.svc.ListAlarmHistory(ctx, params)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.SuccessWithPagination(c, http.StatusOK, result.Alarms, &response.Pagination{
		Page:       result.Page,
		Limit:      result.Limit,
		Total:      result.Total,
		TotalPages: result.TotalPage,
	})
}

func (h *AlarmHandler) AcknowledgeAlarm(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")
	
	// Extract userID from JWT claims
	tokenObj := c.Get("user")
	if tokenObj == nil {
		return response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing auth token")
	}
	token := tokenObj.(*jwt.Token)
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token claims")
	}
	userID := claims["sub"].(string)

	a, err := h.svc.AcknowledgeAlarm(ctx, guid, userID)
	if err != nil {
		if err == alarm.ErrAlarmNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Alarm not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, a, "Alarm acknowledged successfully")
}

func (h *AlarmHandler) ResolveAlarm(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")

	a, err := h.svc.ResolveAlarm(ctx, guid)
	if err != nil {
		if err == alarm.ErrAlarmNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Alarm not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, a, "Alarm resolved successfully")
}

func (h *AlarmHandler) GetAlarmStats(c echo.Context) error {
	ctx := c.Request().Context()
	deviceGUID := c.QueryParam("device_guid")

	var guidPtr *string
	if deviceGUID != "" {
		guidPtr = &deviceGUID
	}

	stats, err := h.svc.GetAlarmStats(ctx, guidPtr)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, stats, "Alarm stats retrieved successfully")
}

func (h *AlarmHandler) StreamAlarms(c echo.Context) error {
	setSSEHeaders(c)

	// Send first data immediately
	if err := h.sendActiveAlarms(c); err != nil {
		writeSSEError(c, err)
		c.Response().Flush()
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request().Context().Done():
			return nil
		case <-ticker.C:
			if err := h.sendActiveAlarms(c); err != nil {
				writeSSEError(c, err)
				c.Response().Flush()
				continue
			}
		}
	}
}

func (h *AlarmHandler) sendActiveAlarms(c echo.Context) error {
	ctx := c.Request().Context()
	// Get all active alarms (limit 50 for performance)
	params := alarm.ListParams{
		Status: strPtr("active"),
		Limit:  50,
		Page:   1,
	}
	result, err := h.svc.ListAlarmHistory(ctx, params)
	if err != nil {
		return err
	}

	return writeSSEData(c, result.Alarms)
}

func strPtr(s string) *string {
	return &s
}
