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
	Create(ctx context.Context, f domain.FamilyItem) (domain.FamilyItem, error)
	Update(ctx context.Context, id int, f domain.FamilyItem) error
	Delete(ctx context.Context, id int) error
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

// familyRequest is the shared inbound shape for POST/PUT, matching legacy's
// $data->Body/$data->Category/$data->Image fields. Image is a pointer:
// legacy never isset()-checks it (optional), unlike Body/Category.
type familyRequest struct {
	Body     string  `json:"Body"`
	Category string  `json:"Category"`
	Image    *string `json:"Image"`
}

// Create handles POST /api/family (auth-gated via router.go), returning
// legacy's 201 status. usecase.FamilyService.Create enforces Body/Category
// presence and the Category enum (see validateFamilyItem).
func (h *FamilyHandler) Create(c *gin.Context) {
	var req familyRequest
	_ = c.ShouldBindJSON(&req)

	f := domain.FamilyItem{Body: req.Body, Category: req.Category, Image: req.Image}
	if _, err := h.svc.Create(c.Request.Context(), f); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Workshops added successfully"})
}

// Update handles PUT /api/family/:id (auth-gated via router.go).
func (h *FamilyHandler) Update(c *gin.Context) {
	id := parseLegacyID(c.Param("id"))
	var req familyRequest
	_ = c.ShouldBindJSON(&req)

	f := domain.FamilyItem{Body: req.Body, Category: req.Category, Image: req.Image}
	if err := h.svc.Update(c.Request.Context(), id, f); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Workshops updated successfully"})
}

// Delete handles DELETE /api/family/:id (auth-gated via router.go). Same
// no-existence-check quirk as legacy.
func (h *FamilyHandler) Delete(c *gin.Context) {
	id := parseLegacyID(c.Param("id"))
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Workshops deleted successfully"})
}
