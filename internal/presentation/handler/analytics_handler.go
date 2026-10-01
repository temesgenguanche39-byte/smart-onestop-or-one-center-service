package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/smart-onestop/platform/internal/application/service"
)

type AnalyticsHandler struct {
	analyticsService *service.AnalyticsService
	slaService       *service.SLAService
}

func NewAnalyticsHandler(
	analyticsService *service.AnalyticsService,
	slaService *service.SLAService,
) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		slaService:       slaService,
	}
}

// GetDashboard returns executive metrics & bottleneck heatmaps
func (h *AnalyticsHandler) GetDashboard(c *gin.Context) {
	var structureID *uint
	if structIDStr := c.Query("structure_id"); structIDStr != "" {
		if id, err := strconv.ParseUint(structIDStr, 10, 32); err == nil {
			u := uint(id)
			structureID = &u
		}
	}

	summary, err := h.analyticsService.GetExecutiveSummary(c.Request.Context(), structureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// TriggerSLASweep allows officials or test runners to trigger immediate SLA breach sweep
func (h *AnalyticsHandler) TriggerSLASweep(c *gin.Context) {
	count, err := h.slaService.SweepAndEscalate(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "SLA escalation sweep executed successfully",
		"escalated_count": count,
	})
}
