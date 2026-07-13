package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/handler/rest/dto"
)

// NewsService is the usecase-layer contract this handler depends on.
type NewsService interface {
	List(ctx context.Context) ([]domain.NewsItem, error)
	Get(ctx context.Context, id int) (domain.NewsItem, error)
}

// NewsHandler serves GET /api/news[/:id], reproducing news.php's contract:
// empty list is 404, by-ID hit is wrapped in a single-element array (see
// dto.NewsResponse doc comment).
type NewsHandler struct {
	svc NewsService
}

func NewNewsHandler(svc NewsService) *NewsHandler {
	return &NewsHandler{svc: svc}
}

func (h *NewsHandler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	if len(items) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No news found"})
		return
	}
	resp := make([]dto.NewsResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, dto.NewNewsResponse(item))
	}
	c.JSON(http.StatusOK, resp)
}

func (h *NewsHandler) Get(c *gin.Context) {
	id := parseLegacyID(c.Param("id"))
	item, err := h.svc.Get(c.Request.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": "No news found"})
		return
	}
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, []dto.NewsResponse{dto.NewNewsResponse(item)})
}
