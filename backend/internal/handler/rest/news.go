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
	Create(ctx context.Context, n domain.NewsItem) (domain.NewsItem, error)
	Update(ctx context.Context, id int, n domain.NewsItem) error
	Delete(ctx context.Context, id int) error
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

// newsRequest is the shared inbound shape for POST/PUT — legacy's news.php
// reads the same 3 fields off the decoded body (PascalCase, matching
// $data->Title etc.). Body and Image are pointers, matching
// domain.NewsItem's nullable-in-schema fields; usecase.NewsService.Create/Update
// rejects a nil/empty value for either, mirroring legacy's isset() check
// on all three (see usecase.validateNewsItem).
type newsRequest struct {
	Title string  `json:"Title"`
	Body  *string `json:"Body"`
	Image *string `json:"Image"`
}

// newsMissingFieldMessage is the exact legacy news.php validation body
// (news.php lines 49-52/80-83: `"Missing Title, Body, or Image parameter"`).
// UNLIKE its siblings (product/gallery/family validation messages, which
// are informal Go-authored text — see e.g. usecase.ProductService.Create's
// "name is required" — since no PR4 requirement demanded verbatim parity
// for THOSE messages), news write validation is required to reproduce this
// string byte-for-byte, so Create/Update special-case domain.ErrValidation
// here instead of delegating to the generic middleware.ErrorHandler body
// (which would otherwise surface usecase.validateNewsItem's Go-idiomatic
// wrapped message text, not the legacy string).
const newsMissingFieldMessage = "Missing Title, Body, or Image parameter"

// Create handles POST /api/news (auth-gated via router.go with the FIXED
// JWT enforcement — see router.go's PR5 doc comment). A malformed/absent
// body binds to a zero-value request (mirrors the null-safe fall-through
// pattern established in AdminHandler.Create/ProductHandler.Create), which
// naturally fails usecase.NewsService.Create's validation as a 400.
func (h *NewsHandler) Create(c *gin.Context) {
	var req newsRequest
	_ = c.ShouldBindJSON(&req)

	n := domain.NewsItem{Title: req.Title, Body: req.Body, Image: req.Image}
	if _, err := h.svc.Create(c.Request.Context(), n); err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"message": newsMissingFieldMessage})
			return
		}
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "News added successfully"})
}

// Update handles PUT /api/news/:id (auth-gated via router.go).
func (h *NewsHandler) Update(c *gin.Context) {
	id := parseLegacyID(c.Param("id"))
	var req newsRequest
	_ = c.ShouldBindJSON(&req)

	n := domain.NewsItem{Title: req.Title, Body: req.Body, Image: req.Image}
	if err := h.svc.Update(c.Request.Context(), id, n); err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"message": newsMissingFieldMessage})
			return
		}
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "News updated successfully"})
}

// Delete handles DELETE /api/news/:id (auth-gated via router.go). Same
// no-existence-check quirk as legacy deleteNews.
func (h *NewsHandler) Delete(c *gin.Context) {
	id := parseLegacyID(c.Param("id"))
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "News deleted successfully"})
}
