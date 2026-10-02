package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smart-onestop/platform/internal/application/dto"
	"github.com/smart-onestop/platform/internal/application/service"
	"github.com/smart-onestop/platform/internal/domain"
	"github.com/smart-onestop/platform/internal/domain/workcalendar"
	"github.com/smart-onestop/platform/internal/presentation/middleware"
)

type HearingHandler struct {
	hearingService *service.HearingService
	hearingRepo    domain.HearingRepository
}

func NewHearingHandler(hearingService *service.HearingService, hearingRepo domain.HearingRepository) *HearingHandler {
	return &HearingHandler{
		hearingService: hearingService,
		hearingRepo:    hearingRepo,
	}
}

// ScheduleSlot for case
func (h *HearingHandler) ScheduleSlot(c *gin.Context) {
	idStr := c.Param("id")
	caseID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case UUID"})
		return
	}

	var req dto.ScheduleHearingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	officialID := middleware.GetCurrentUserID(c)
	resp, err := h.hearingService.ScheduleSlot(c.Request.Context(), caseID, officialID, req)
	if err != nil {
		var holErr *service.HearingHolidayError
		if errors.As(err, &holErr) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":                         holErr.Error(),
				"holiday_date":                  holErr.Date.Format("2006-01-02"),
				"suggested_next_slot":           holErr.SuggestedSlot.Format("2006-01-02"),
				"suggested_next_slot_ethiopian": workcalendar.FormatEthiopian(workcalendar.ToEthiopianDate(holErr.SuggestedSlot), "am"),
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ListByDate gets hearings for a specific date
func (h *HearingHandler) ListByDate(c *gin.Context) {
	dateStr := c.Param("date")
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, expected YYYY-MM-DD"})
		return
	}

	var officialID *uuid.UUID
	if offStr := c.Query("official_id"); offStr != "" {
		if id, err := uuid.Parse(offStr); err == nil {
			officialID = &id
		}
	}

	slots, err := h.hearingRepo.ListByDate(c.Request.Context(), parsedDate, officialID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, slots)
}

// CompleteHearing marks slot completed
func (h *HearingHandler) CompleteHearing(c *gin.Context) {
	idStr := c.Param("id")
	slotID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hearing slot UUID"})
		return
	}

	var req struct {
		HearingNotes string `json:"hearing_notes" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	officialID := middleware.GetCurrentUserID(c)
	if err := h.hearingService.CompleteHearing(c.Request.Context(), slotID, officialID, req.HearingNotes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Hearing marked completed"})
}
