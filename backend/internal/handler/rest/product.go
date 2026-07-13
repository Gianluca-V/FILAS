package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/handler/rest/dto"
)

// ProductService is the usecase-layer contract this handler depends on.
type ProductService interface {
	List(ctx context.Context) ([]domain.Product, error)
	Get(ctx context.Context, id int) (domain.Product, error)
}

// ProductHandler serves GET /api/products[/:id], reproducing
// products.php's contract. UNLIKE the other four public resources, this
// one does NOT share their generic quirks: an empty list is 200 [] (not
// 404), and a found item is a bare JSON object (not array-wrapped) —
// products.php has its own dedicated getProducts()/getProduct() functions
// instead of reusing the others' shared row-collector code path.
type ProductHandler struct {
	svc ProductService
}

func NewProductHandler(svc ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

func (h *ProductHandler) List(c *gin.Context) {
	products, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	resp := make([]dto.ProductResponse, 0, len(products))
	for _, p := range products {
		resp = append(resp, dto.NewProductResponse(p))
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ProductHandler) Get(c *gin.Context) {
	id := parseLegacyID(c.Param("id"))
	product, err := h.svc.Get(c.Request.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Product not found"})
		return
	}
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.NewProductResponse(product))
}
