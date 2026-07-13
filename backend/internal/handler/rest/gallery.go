package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/handler/rest/dto"
)

// GalleryService is the usecase-layer contract this handler depends on.
type GalleryService interface {
	List(ctx context.Context) ([]domain.GalleryImage, error)
	Get(ctx context.Context, id int) (domain.GalleryImage, error)
}

// GalleryHandler serves GET /api/gallery[/:id], reproducing gallery.php's
// contract: an empty list is 404 (not 200 []), and a by-ID hit is wrapped
// in a single-element array (see dto.GalleryResponse doc comment).
type GalleryHandler struct {
	svc GalleryService
}

func NewGalleryHandler(svc GalleryService) *GalleryHandler {
	return &GalleryHandler{svc: svc}
}

func (h *GalleryHandler) List(c *gin.Context) {
	images, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	if len(images) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No images found"})
		return
	}
	resp := make([]dto.GalleryResponse, 0, len(images))
	for _, img := range images {
		resp = append(resp, dto.NewGalleryResponse(img))
	}
	c.JSON(http.StatusOK, resp)
}

func (h *GalleryHandler) Get(c *gin.Context) {
	id := parseLegacyID(c.Param("id"))
	img, err := h.svc.Get(c.Request.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": "No images found"})
		return
	}
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, []dto.GalleryResponse{dto.NewGalleryResponse(img)})
}
