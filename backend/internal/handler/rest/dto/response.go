// Package dto holds outbound response shapes matching the legacy PHP JSON
// contract exactly, including its quirks (see per-type doc comments and
// backend/docs/legacy-quirks.md): numeric columns are encoded as JSON
// strings because that is what the legacy mysqli-backed PHP API actually
// emits (characterized live against FilasServer/*.php + the seeded DB).
// This is deliberate contract fidelity, not a modeling mistake.
package dto

import (
	"strconv"
	"time"

	"github.com/gianluca-v/filas-backend/internal/domain"
)

// ProductResponse matches GET /api/products[/:id] from products.php, which
// returns a single JSON OBJECT for a by-ID lookup (unlike the other four
// resources below, which wrap a by-ID result in a single-element array —
// see NewsResponse etc.). Numeric fields are JSON strings; Description and
// Image are both nullable and preserved as null vs "" exactly as stored
// (see gate #34: a NULL Image column used to 500 the whole endpoint).
type ProductResponse struct {
	ID          string  `json:"ID"`
	Name        string  `json:"Name"`
	Price       string  `json:"Price"`
	Stock       string  `json:"Stock"`
	Image       *string `json:"Image"`
	Description *string `json:"Description"`
}

// NewProductResponse converts a domain.Product to its legacy-shaped JSON
// response. formatLegacyFloat mirrors MySQL's double->text formatting
// (shortest round-trip, no forced trailing zeros).
func NewProductResponse(p domain.Product) ProductResponse {
	return ProductResponse{
		ID:          strconv.Itoa(p.ID),
		Name:        p.Name,
		Price:       formatLegacyFloat(p.Price),
		Stock:       strconv.Itoa(p.Stock),
		Image:       p.Image,
		Description: p.Description,
	}
}

func formatLegacyFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// NewsResponse matches one element of GET /api/news[/:id] from news.php.
// CRITICAL QUIRK: a by-ID lookup in the legacy code reuses the exact same
// row-collecting loop as the list endpoint, so GET /api/news/:id returns a
// JSON ARRAY with one element, NOT a bare object. Handlers must wrap the
// single result in a slice to match (see rest.NewsHandler.Get).
type NewsResponse struct {
	ID    string  `json:"ID"`
	Title string  `json:"Title"`
	Body  *string `json:"Body"`
	Image *string `json:"Image"`
}

func NewNewsResponse(n domain.NewsItem) NewsResponse {
	return NewsResponse{ID: strconv.Itoa(n.ID), Title: n.Title, Body: n.Body, Image: n.Image}
}

// GalleryResponse matches one element of GET /api/gallery[/:id] from
// gallery.php. Same by-ID array-wrap quirk as NewsResponse.
type GalleryResponse struct {
	ID    string `json:"ID"`
	Image string `json:"Image"`
}

func NewGalleryResponse(g domain.GalleryImage) GalleryResponse {
	return GalleryResponse{ID: strconv.Itoa(g.ID), Image: g.Image}
}

// FamilyResponse matches one element of GET /api/family[/:id] from
// family.php. Same by-ID array-wrap quirk as NewsResponse.
type FamilyResponse struct {
	ID       string  `json:"ID"`
	Image    *string `json:"Image"`
	Body     string  `json:"Body"`
	Category string  `json:"Category"`
}

func NewFamilyResponse(f domain.FamilyItem) FamilyResponse {
	return FamilyResponse{ID: strconv.Itoa(f.ID), Image: f.Image, Body: f.Body, Category: f.Category}
}

// OrganizationResponse matches one element of GET /api/organizations[/:id]
// from organizations.php. Same by-ID array-wrap quirk as NewsResponse.
type OrganizationResponse struct {
	ID          string  `json:"ID"`
	Title       string  `json:"Title"`
	Description *string `json:"Description"`
	Image       string  `json:"Image"`
}

func NewOrganizationResponse(o domain.Organization) OrganizationResponse {
	return OrganizationResponse{ID: strconv.Itoa(o.ID), Title: o.Title, Description: o.Description, Image: o.Image}
}

// AdminResponse matches one element of GET /api/admins[/:id] from
// admins.php, MINUS password and salt. Legacy used `SELECT *` and echoed
// the raw row — including the password hash and salt — in the JSON body
// (characterized live, see backend/docs/legacy-quirks.md §1). This is a
// deliberate security fix, not preserved: AdminResponse structurally has
// no Password/Salt field to serialize. The JSON key is lowercase
// "username" (not "Username"), matching the `username` column name legacy
// echoed verbatim via fetch_assoc.
type AdminResponse struct {
	ID       string `json:"ID"`
	Username string `json:"username"`
}

func NewAdminResponse(a domain.Admin) AdminResponse {
	return AdminResponse{ID: strconv.Itoa(a.ID), Username: a.Username}
}

// LoginResponse matches POST /api/admins {login:true} on success.
type LoginResponse struct {
	Token string `json:"token"`
}

// legacyDateTimeLayout matches MySQL's DATETIME string form, exactly what
// legacy's mysqli fetch_assoc echoed verbatim for orders.startDate/
// finishDate (characterized live: `"2023-11-20 18:30:02"`, no timezone
// offset, no "T" separator — see backend/docs/legacy-quirks.md §15).
const legacyDateTimeLayout = "2006-01-02 15:04:05"

// OrderProductResponse is one line item inside OrderResponse.Products.
// FIELD ORDER AND TYPES ARE LOAD-BEARING: this reproduces the shape
// legacy's `json_decode(stripslashes($row['products']), true)` produced
// from its GROUP_CONCAT-built JSON string (orders.php:83-93/133-143),
// where ProductPrice/ProductQuantity are genuine JSON NUMBERS (unquoted —
// unlike every other numeric field in this API, which is a JSON STRING,
// see e.g. ProductResponse), because they came from json_decode-ing a
// hand-built JSON string, not from a raw mysqli string column.
// CRITICAL, live-confirmed characterization: ProductPrice is the
// PRODUCT'S CURRENT price (`p.price` in a live JOIN), NOT the frozen
// per-line price actually billed (`op.orderPrice` = price*quantity at
// order time) — see domain.OrderProduct.ProductName's doc comment.
type OrderProductResponse struct {
	ProductName     string  `json:"productName"`
	ProductPrice    float64 `json:"productPrice"`
	ProductQuantity int     `json:"productQuantity"`
}

// OrderResponse matches one element of GET /api/orders[/:id] from
// orders.php, characterized live against a scratch copy of FilasServer/
// running against the seeded Docker db (see backend/docs/legacy-quirks.md
// §15 for the full capture). FIELD ORDER MATTERS (Go's encoding/json
// preserves struct declaration order): orderID, orderTotal, orderState,
// orderStartDate, orderFinishDate, orderName, orderPhone, products — this
// is the EXACT order legacy's associative-array-turned-JSON produced.
// Unlike NewsResponse/GalleryResponse/etc., a by-ID GET also wraps the
// single result in an array (see rest.OrderHandler.Get), matching the
// SAME array-wrap quirk (§8) for a different underlying reason: legacy's
// getOrder collects rows into `$order[]` even though the WHERE clause can
// only ever match one order's worth of JOINed rows.
//
// Deliberately NOT reproduced: legacy's manual `CONCAT(...)`-built
// `products` JSON string is fragile — a product Name containing a `"` or
// `,` corrupts the concatenated string, and `json_decode` on the corrupt
// result silently returns null, so legacy emits `"products":null` for
// such an order (live-confirmed: see legacy-quirks.md §15). Go's
// `encoding/json` builds this array structurally and properly escapes
// quotes/commas, so `products` is ALWAYS a valid array — reproducing the
// SHAPE, not the bug's fragility, per the design's own precedent (ADR
// notes on the GROUP_CONCAT risk).
type OrderResponse struct {
	ID         string                 `json:"orderID"`
	Total      string                 `json:"orderTotal"`
	State      string                 `json:"orderState"`
	StartDate  string                 `json:"orderStartDate"`
	FinishDate *string                `json:"orderFinishDate"`
	Name       string                 `json:"orderName"`
	Phone      string                 `json:"orderPhone"`
	Products   []OrderProductResponse `json:"products"`
}

// NewOrderResponse converts a domain.Order (as returned by
// OrderRepository.List/Get, with ProductName/ProductPrice populated) to
// its legacy-shaped JSON response.
func NewOrderResponse(o domain.Order) OrderResponse {
	products := make([]OrderProductResponse, 0, len(o.Products))
	for _, p := range o.Products {
		products = append(products, OrderProductResponse{
			ProductName:     p.ProductName,
			ProductPrice:    p.ProductPrice,
			ProductQuantity: p.Quantity,
		})
	}
	resp := OrderResponse{
		ID:        strconv.Itoa(o.ID),
		Total:     formatLegacyFloat(o.Total),
		State:     string(o.State),
		StartDate: formatLegacyDateTime(o.StartDate),
		Name:      o.Name,
		Phone:     o.Phone,
		Products:  products,
	}
	if o.FinishDate != nil {
		formatted := formatLegacyDateTime(*o.FinishDate)
		resp.FinishDate = &formatted
	}
	return resp
}

func formatLegacyDateTime(t time.Time) string {
	return t.Format(legacyDateTimeLayout)
}
