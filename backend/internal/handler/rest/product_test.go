package rest_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/handler/rest"
)

type fakeProductService struct {
	products []domain.Product
	byID     map[int]domain.Product
	err      error
}

func (f fakeProductService) List(ctx context.Context) ([]domain.Product, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.products, nil
}

func (f fakeProductService) Get(ctx context.Context, id int) (domain.Product, error) {
	if f.err != nil {
		return domain.Product{}, f.err
	}
	p, ok := f.byID[id]
	if !ok {
		return domain.Product{}, domain.ErrNotFound
	}
	return p, nil
}

func newProductTestRouter(svc rest.ProductService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := rest.NewProductHandler(svc)
	r.GET("/api/products", h.List)
	r.GET("/api/products/:id", h.Get)
	return r
}

func TestProductHandler_List_ReturnsArray(t *testing.T) {
	svc := fakeProductService{products: []domain.Product{
		{ID: 1, Name: "Mermelada de pera", Price: 600, Stock: 112, Image: "assets/default-img.png"},
	}}
	r := newProductTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	want := `[{"ID":"1","Name":"Mermelada de pera","Price":"600","Stock":"112","Image":"assets/default-img.png","Description":null}]`
	if rec.Body.String() != want {
		t.Errorf("body = %s, want %s", rec.Body.String(), want)
	}
}

func TestProductHandler_List_ReturnsEmptyArrayWhenNoProducts(t *testing.T) {
	// Legacy quirk (characterized live, obs #29/#31): unlike the other four
	// resources, products.php returns 200 [] on an empty table, NOT 404.
	// products.php has dedicated getProducts()/getProduct() functions that
	// don't share the generic "num_rows==0 -> 404" code path the others do.
	r := newProductTestRouter(fakeProductService{products: []domain.Product{}})

	req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Body.String() != `[]` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `[]`)
	}
}

func TestProductHandler_Get_ReturnsBareObjectNotArray(t *testing.T) {
	// Legacy quirk: unlike news/gallery/family/organizations, a found
	// product is a bare JSON object, not wrapped in an array.
	svc := fakeProductService{byID: map[int]domain.Product{1: {ID: 1, Name: "Mermelada de pera", Price: 600, Stock: 112, Image: "assets/default-img.png"}}}
	r := newProductTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/products/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	want := `{"ID":"1","Name":"Mermelada de pera","Price":"600","Stock":"112","Image":"assets/default-img.png","Description":null}`
	if rec.Body.String() != want {
		t.Errorf("body = %s, want %s", rec.Body.String(), want)
	}
}

func TestProductHandler_Get_ReturnsNotFoundForMissingID(t *testing.T) {
	r := newProductTestRouter(fakeProductService{byID: map[int]domain.Product{}})

	req := httptest.NewRequest(http.MethodGet, "/api/products/999999", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if rec.Body.String() != `{"message":"Product not found"}` {
		t.Errorf("body = %s, want %s", rec.Body.String(), `{"message":"Product not found"}`)
	}
}

func TestProductHandler_Get_TreatsNonNumericIDAsNotFound(t *testing.T) {
	r := newProductTestRouter(fakeProductService{byID: map[int]domain.Product{}})

	req := httptest.NewRequest(http.MethodGet, "/api/products/abc", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}
