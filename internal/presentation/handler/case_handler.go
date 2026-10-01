package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smart-onestop/platform/internal/application/dto"
	"github.com/smart-onestop/platform/internal/application/service"
	"github.com/smart-onestop/platform/internal/domain"
	"github.com/smart-onestop/platform/internal/presentation/middleware"
)

type CaseHandler struct {
	caseService *service.CaseService
}

func NewCaseHandler(caseService *service.CaseService) *CaseHandler {
	return &CaseHandler{caseService: caseService}
}

// CreateCase (Intake)
func (h *CaseHandler) CreateCase(c *gin.Context) {
	var req dto.CreateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.caseService.CreateCase(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetByTicket public tracking
func (h *CaseHandler) GetByTicket(c *gin.Context) {
	ticket := strings.TrimSpace(c.Param("ticket"))
	if ticket == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ticket number is required"})
		return
	}

	if strings.HasPrefix(ticket, "ETH-MUNI-") {
		resp, err := h.caseService.GetByQR(c.Request.Context(), ticket)
		if err == nil {
			c.JSON(http.StatusOK, resp)
			return
		}
	}

	resp, err := h.caseService.GetByTicket(c.Request.Context(), ticket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// VerifyQR public cryptographic verification
func (h *CaseHandler) VerifyQR(c *gin.Context) {
	qrCode := strings.TrimSpace(c.Param("qr"))
	if qrCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "QR verification code is required"})
		return
	}

	resp, err := h.caseService.GetByQR(c.Request.Context(), qrCode)
	if err != nil {
		if strings.HasPrefix(qrCode, "TKT-") {
			resp, err = h.caseService.GetByTicket(c.Request.Context(), qrCode)
		}
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid or unverified QR code"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListCases for officials
func (h *CaseHandler) ListCases(c *gin.Context) {
	filter := domain.CaseFilter{
		SearchQuery: c.Query("search"),
	}

	if structIDStr := c.Query("structure_id"); structIDStr != "" {
		if id, err := strconv.ParseUint(structIDStr, 10, 32); err == nil {
			u := uint(id)
			filter.StructureID = &u
		}
	}

	if statusStr := c.Query("status"); statusStr != "" {
		s := domain.CaseStatus(statusStr)
		filter.Status = &s
	}

	if priorityStr := c.Query("priority"); priorityStr != "" {
		p := domain.Priority(priorityStr)
		filter.Priority = &p
	}

	if breachedStr := c.Query("only_breached"); breachedStr == "true" {
		filter.OnlyBreached = true
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = o
		}
	}

	cases, total, err := h.caseService.ListCases(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   cases,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetByID detailed view
func (h *CaseHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case UUID"})
		return
	}

	resp, err := h.caseService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AssignCase to official
func (h *CaseHandler) AssignCase(c *gin.Context) {
	idStr := c.Param("id")
	caseID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case UUID"})
		return
	}

	var req dto.AssignCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	performerID := middleware.GetCurrentUserID(c)
	if err := h.caseService.AssignCase(c.Request.Context(), caseID, req.AssignedToUserID, performerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Case assigned successfully"})
}

// ResolveCase with official resolution
func (h *CaseHandler) ResolveCase(c *gin.Context) {
	idStr := c.Param("id")
	caseID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case UUID"})
		return
	}

	var req dto.ResolveCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	officialID := middleware.GetCurrentUserID(c)
	if err := h.caseService.ResolveCase(c.Request.Context(), caseID, officialID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Case resolved successfully and decision certified"})
}

// ManualEscalate for supervisors
func (h *CaseHandler) ManualEscalate(c *gin.Context) {
	idStr := c.Param("id")
	caseID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case UUID"})
		return
	}

	var req dto.ManualEscalateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	performerID := middleware.GetCurrentUserID(c)
	if err := h.caseService.ManualEscalate(c.Request.Context(), caseID, performerID, req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Case escalated to higher administrative tier"})
}

// AddAttachment handles official document and audio memo uploads
func (h *CaseHandler) AddAttachment(c *gin.Context) {
	idStr := c.Param("id")
	caseID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case UUID format"})
		return
	}

	var req dto.AddAttachmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	officialID := middleware.GetCurrentUserID(c)
	att, err := h.caseService.AddAttachment(c.Request.Context(), caseID, &officialID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, att)
}
