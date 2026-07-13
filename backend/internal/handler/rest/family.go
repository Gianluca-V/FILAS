package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/handler/rest/dto"
)

// FamilyService is the usecase-layer contract this handler depends on.
type FamilyService interface {
	List(ctx context.Context) ([]domain.FamilyItem, error)
	Get(ctx context.Context, id int) (domain.FamilyItem, error)
}

// FamilyHandler serves GET /api/family[/:id], reproducing family.php's
// contract: empty list is 404 ("No workshops found" — the legacy code's
// internal naming for this resource), by-ID hit is wrapped in a
// single-element array (see dto.FamilyResponse doc comment).
type FamilyHandler struct {
	svc FamilyService
}

func NewFamilyHandler(svc FamilyService) *FamilyHandler {
	return &FamilyHandler{svc: svc}
}

func (h *FamilyHandler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	if len(items) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No workshops found"})
		return
	}
	resp := make([]dto.FamilyResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, dto.NewFamilyResponse(item))
	}
	c.JSON(http.StatusOK, resp)
}

func (h *FamilyHandler) Get(c *gin.Context) {
	id := parseLegacyID(c.Param("id"))
	item, err := h.svc.Get(c.Request.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": "No workshops found"})
		return
	}
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, []dto.FamilyResponse{dto.NewFamilyResponse(item)})
}
