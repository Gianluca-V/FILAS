package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/handler/rest/dto"
)

// OrganizationService is the usecase-layer contract this handler depends on.
type OrganizationService interface {
	List(ctx context.Context) ([]domain.Organization, error)
	Get(ctx context.Context, id int) (domain.Organization, error)
}

// OrganizationHandler serves GET /api/organizations[/:id], reproducing
// organizations.php's contract: empty list is 404, by-ID hit is wrapped in
// a single-element array (see dto.OrganizationResponse doc comment).
type OrganizationHandler struct {
	svc OrganizationService
}

func NewOrganizationHandler(svc OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{svc: svc}
}

func (h *OrganizationHandler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	if len(items) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No organizations found"})
		return
	}
	resp := make([]dto.OrganizationResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, dto.NewOrganizationResponse(item))
	}
	c.JSON(http.StatusOK, resp)
}

func (h *OrganizationHandler) Get(c *gin.Context) {
	id := parseLegacyID(c.Param("id"))
	item, err := h.svc.Get(c.Request.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": "No organizations found"})
		return
	}
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, []dto.OrganizationResponse{dto.NewOrganizationResponse(item)})
}
