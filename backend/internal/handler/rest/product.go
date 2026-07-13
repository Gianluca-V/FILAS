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
	Create(ctx context.Context, p domain.Product) (domain.Product, error)
	Update(ctx context.Context, id int, p domain.Product) error
	Delete(ctx context.Context, id int) error
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

// productRequest is the shared inbound shape for POST/PUT — legacy's
// createProduct and updateProduct both read the same 5 fields off the
// decoded body (PascalCase, matching $data->Name etc. in products.php).
type productRequest struct {
	Name        string  `json:"Name"`
	Price       float64 `json:"Price"`
	Stock       int     `json:"Stock"`
	Image       *string `json:"Image"`
	Description *string `json:"Description"`
}

// Create handles POST /api/products (auth-gated via router.go). Legacy
// createProduct has NO field validation; usecase.ProductService.Create adds
// a deliberate Name-required check (see its doc comment). A
// malformed/absent body binds to a zero-value request (mirrors the
// null-safe fall-through pattern established in AdminHandler.Create),
// which naturally fails that validation as a 400.
func (h *ProductHandler) Create(c *gin.Context) {
	var req productRequest
	_ = c.ShouldBindJSON(&req)

	p := domain.Product{Name: req.Name, Price: req.Price, Stock: req.Stock, Image: req.Image, Description: req.Description}
	if _, err := h.svc.Create(c.Request.Context(), p); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product created successfully"})
}

// Update handles PUT /api/products/:id (auth-gated via router.go). Like
// legacy updateProduct, Description is never persisted here (see
// domain.ProductRepository.Update).
func (h *ProductHandler) Update(c *gin.Context) {
	id := parseLegacyID(c.Param("id"))
	var req productRequest
	_ = c.ShouldBindJSON(&req)

	p := domain.Product{Name: req.Name, Price: req.Price, Stock: req.Stock, Image: req.Image, Description: req.Description}
	if err := h.svc.Update(c.Request.Context(), id, p); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
}

// Delete handles DELETE /api/products/:id (auth-gated via router.go). Same
// no-existence-check quirk as legacy deleteProduct.
func (h *ProductHandler) Delete(c *gin.Context) {
	id := parseLegacyID(c.Param("id"))
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
