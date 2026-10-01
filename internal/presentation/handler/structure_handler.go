package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smart-onestop/platform/internal/domain"
)

type StructureHandler struct {
	structureRepo   domain.StructureRepository
	serviceTypeRepo domain.ServiceTypeRepository
}

func NewStructureHandler(
	structureRepo domain.StructureRepository,
	serviceTypeRepo domain.ServiceTypeRepository,
) *StructureHandler {
	return &StructureHandler{
		structureRepo:   structureRepo,
		serviceTypeRepo: serviceTypeRepo,
	}
}

// GetTree returns administrative hierarchy tree (City -> Sub-Cities -> Woredas)
func (h *StructureHandler) GetTree(c *gin.Context) {
	tree, err := h.structureRepo.GetHierarchyTree(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tree)
}

// ListAll returns flat list of all municipal structures
func (h *StructureHandler) ListAll(c *gin.Context) {
	list, err := h.structureRepo.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// ListServiceTypes returns service types with base SLA hours
func (h *StructureHandler) ListServiceTypes(c *gin.Context) {
	list, err := h.serviceTypeRepo.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}
