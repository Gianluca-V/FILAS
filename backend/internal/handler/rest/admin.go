package rest

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/domain"
	"github.com/gianluca-v/filas-backend/internal/handler/rest/dto"
)

// AdminService is the usecase-layer contract this handler depends on for
// admin CRUD (login lives separately, see AuthService below).
type AdminService interface {
	List(ctx context.Context) ([]domain.Admin, error)
	Get(ctx context.Context, id int) (domain.Admin, error)
	Create(ctx context.Context, username, password string) (domain.Admin, error)
	Update(ctx context.Context, id int, username, password string) error
	Delete(ctx context.Context, id int) error
}

// AuthService is the usecase-layer contract for admin login.
type AuthService interface {
	Login(ctx context.Context, username, password string) (string, error)
}

// tokenParser is satisfied by *auth.JWTService; kept narrow for testability
// (same shape as middleware.jwtParser, duplicated here because that one is
// package-private — see AdminHandler.requireAuthInline).
type tokenParser interface {
	Parse(tokenString string) (int64, error)
}

// AdminHandler serves /api/admins, reproducing admins.php's contract.
// Unlike the other resources, POST is NOT auth-gated by route-level
// middleware: legacy branches on the request body itself (login vs
// create), so the auth check for the non-login case is performed INLINE
// after decoding the body once (see Create and requireAuthInline).
// GET/PUT/DELETE are gated normally via middleware.RequireAuth in
// router.go, mirroring legacy's uniform requirement there.
type AdminHandler struct {
	adminSvc AdminService
	authSvc  AuthService
	jwt      tokenParser
}

// NewAdminHandler wires an AdminHandler to its usecases and JWT parser.
func NewAdminHandler(adminSvc AdminService, authSvc AuthService, jwt tokenParser) *AdminHandler {
	return &AdminHandler{adminSvc: adminSvc, authSvc: authSvc, jwt: jwt}
}

// requireAuthInline mirrors middleware.RequireAuth's check for the one
// route (POST /api/admins, non-login branch) that cannot use route-group
// middleware because the same path/method also serves the public login
// call. Returns false (having already written the 401 response) if the
// request is not authenticated.
func (h *AdminHandler) requireAuthInline(c *gin.Context) bool {
	header := c.GetHeader("Authorization")
	if header == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "missing authorization token"})
		return false
	}
	token := strings.TrimPrefix(header, "Bearer ")
	if _, err := h.jwt.Parse(token); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid or expired token"})
		return false
	}
	return true
}

type createAdminRequest struct {
	Login    bool   `json:"login"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Create handles POST /api/admins, which legacy overloads for two
// purposes based on the body: {"login":true,...} performs an
// unauthenticated login (checkUserForLogin), anything else creates a new
// admin and requires a valid JWT (createUser). A malformed/absent body is
// treated as an empty request (mirrors PHP's null-safe isset() checks on
// json_decode(...) === null), which naturally falls through to the
// appropriate 400/401.
func (h *AdminHandler) Create(c *gin.Context) {
	var req createAdminRequest
	_ = c.ShouldBindJSON(&req)

	if req.Login {
		token, err := h.authSvc.Login(c.Request.Context(), req.Username, req.Password)
		if errors.Is(err, domain.ErrUnauthorized) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Login failed"})
			return
		}
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, dto.LoginResponse{Token: token})
		return
	}

	if !h.requireAuthInline(c) {
		return
	}

	// Unlike legacy (which also required a caller-supplied ID), only
	// username/password are required here — the seed schema's
	// AUTO_INCREMENT PRIMARY KEY assigns the ID (task 2.1).
	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing username or password parameters"})
		return
	}

	if _, err := h.adminSvc.Create(c.Request.Context(), req.Username, req.Password); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
}

// List handles GET /api/admins (auth-gated via router.go). Legacy: 200 []
// when empty (NOT 404 — same pattern as products.php, unlike the other
// four public resources).
func (h *AdminHandler) List(c *gin.Context) {
	admins, err := h.adminSvc.List(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	resp := make([]dto.AdminResponse, 0, len(admins))
	for _, a := range admins {
		resp = append(resp, dto.NewAdminResponse(a))
	}
	c.JSON(http.StatusOK, resp)
}

// Get handles GET /api/admins/:id (auth-gated via router.go). Legacy
// getUser() returns a bare object, not array-wrapped.
func (h *AdminHandler) Get(c *gin.Context) {
	id := parseLegacyID(c.Param("id"))
	admin, err := h.adminSvc.Get(c.Request.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.NewAdminResponse(admin))
}

type updateAdminRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Update handles PUT /api/admins/:id (auth-gated via router.go). Bcrypt
// hashes the new password (fixes the legacy floatval($password) bug — see
// usecase.AdminService.Update). Like legacy, it does not check whether the
// ID exists first (always reports success).
func (h *AdminHandler) Update(c *gin.Context) {
	id := parseLegacyID(c.Param("id"))
	var req updateAdminRequest
	_ = c.ShouldBindJSON(&req)

	if err := h.adminSvc.Update(c.Request.Context(), id, req.Username, req.Password); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// Delete handles DELETE /api/admins/:id (auth-gated via router.go). Same
// no-existence-check quirk as Update, preserved for contract fidelity.
func (h *AdminHandler) Delete(c *gin.Context) {
	id := parseLegacyID(c.Param("id"))
	if err := h.adminSvc.Delete(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
