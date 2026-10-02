package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/smart-onestop/platform/internal/application/dto"
	"github.com/smart-onestop/platform/internal/application/service"
)

type CalendarHandler struct {
	calendarService *service.CalendarService
}

func NewCalendarHandler(calendarService *service.CalendarService) *CalendarHandler {
	return &CalendarHandler{calendarService: calendarService}
}

// GetWorkingDays handles GET /api/v1/calendar/working-days?from=&to=&calendar=ethiopian|gregorian
func (h *CalendarHandler) GetWorkingDays(c *gin.Context) {
	var req dto.WorkingDaysRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameters 'from' and 'to' are required in YYYY-MM-DD format"})
		return
	}

	resp, err := h.calendarService.GetWorkingDays(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetHolidays handles GET /api/v1/calendar/holidays?year=
func (h *CalendarHandler) GetHolidays(c *gin.Context) {
	yearStr := c.Query("year")
	year := 0
	if yearStr != "" {
		if parsed, err := strconv.Atoi(yearStr); err == nil {
			year = parsed
		}
	}

	holidays, err := h.calendarService.GetHolidays(c.Request.Context(), year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"year":     year,
		"count":    len(holidays),
		"holidays": holidays,
	})
}

// GetNextHearingSlots handles GET /api/v1/calendar/next-hearing-slots?from=&count=
func (h *CalendarHandler) GetNextHearingSlots(c *gin.Context) {
	from := c.Query("from")
	countStr := c.Query("count")
	count := 5
	if countStr != "" {
		if parsed, err := strconv.Atoi(countStr); err == nil && parsed > 0 {
			count = parsed
		}
	}

	resp, err := h.calendarService.GetNextHearingSlots(c.Request.Context(), from, count)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CreateHoliday handles POST /api/v1/calendar/holidays and POST /holidays (SUPER_ADMIN only)
func (h *CalendarHandler) CreateHoliday(c *gin.Context) {
	var req dto.CreateHolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.calendarService.CreateHoliday(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// UpdateHoliday handles PUT /api/v1/calendar/holidays/:id and PUT /holidays/:id (SUPER_ADMIN only)
func (h *CalendarHandler) UpdateHoliday(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid holiday ID"})
		return
	}

	var req dto.UpdateHolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.calendarService.UpdateHoliday(c.Request.Context(), uint(id), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteHoliday handles DELETE /api/v1/calendar/holidays/:id and DELETE /holidays/:id (SUPER_ADMIN only)
func (h *CalendarHandler) DeleteHoliday(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid holiday ID"})
		return
	}

	if err := h.calendarService.DeleteHoliday(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Holiday successfully deleted", "id": id})
}
