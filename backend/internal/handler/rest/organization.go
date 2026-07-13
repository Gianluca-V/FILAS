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
	Create(ctx context.Context, o domain.Organization) (domain.Organization, error)
	Update(ctx context.Context, id int, o domain.Organization) error
	Delete(ctx context.Context, id int) error
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

// organizationRequest is the shared inbound shape for POST/PUT, matching
// legacy's $data->Title/$data->Description/$data->Image fields. Description
// is a pointer: legacy never isset()-checks it (optional), unlike
// Title/Image (see usecase.validateOrganization).
type organizationRequest struct {
	Title       string  `json:"Title"`
	Description *string `json:"Description"`
	Image       string  `json:"Image"`
}

// Create handles POST /api/organizations (auth-gated via router.go),
// returning legacy's 201 status and message. usecase.OrganizationService.Create
// enforces Title/Image presence (see validateOrganization). A
// malformed/absent body binds to a zero-value request (mirrors the
// null-safe fall-through pattern established in AdminHandler.Create), which
// naturally fails that validation as a 400.
func (h *OrganizationHandler) Create(c *gin.Context) {
	var req organizationRequest
	_ = c.ShouldBindJSON(&req)

	o := domain.Organization{Title: req.Title, Description: req.Description, Image: req.Image}
	if _, err := h.svc.Create(c.Request.Context(), o); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Organizations added successfully"})
}

// Update handles PUT /api/organizations/:id (auth-gated via router.go).
func (h *OrganizationHandler) Update(c *gin.Context) {
	id := parseLegacyID(c.Param("id"))
	var req organizationRequest
	_ = c.ShouldBindJSON(&req)

	o := domain.Organization{Title: req.Title, Description: req.Description, Image: req.Image}
	if err := h.svc.Update(c.Request.Context(), id, o); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Organizations updated successfully"})
}

// Delete handles DELETE /api/organizations/:id (auth-gated via router.go).
// Same no-existence-check quirk as legacy. The lowercase "organizations" in
// the success message (unlike Create/Update's capitalized "Organizations")
// is a live-characterized legacy quirk (organizations.php's delete branch
// uses a different literal than create/update), preserved deliberately for
// contract fidelity, not a typo.
func (h *OrganizationHandler) Delete(c *gin.Context) {
	id := parseLegacyID(c.Param("id"))
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "organizations deleted successfully"})
}
