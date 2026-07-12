// Package auth provides token issuance/verification (JWT) and password
// hashing/verification (bcrypt + legacy sha256 migration) primitives used
// by the usecase and handler layers.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// claims mirrors the legacy PHP payload shape ({exp, data: adminID}) for
// contract parity, plus a standard "iat".
type claims struct {
	Data int64 `json:"data"`
	jwt.RegisteredClaims
}

// JWTService issues and verifies HS256 JWTs carrying an admin ID.
type JWTService struct {
	secret []byte
	expiry time.Duration
}

// NewJWTService builds a JWTService with the given signing secret and
// token expiry duration.
func NewJWTService(secret string, expiry time.Duration) *JWTService {
	return &JWTService{secret: []byte(secret), expiry: expiry}
}

// Generate issues a signed JWT carrying adminID, expiring after the
// service's configured duration.
func (s *JWTService) Generate(adminID int64) (string, error) {
	now := time.Now()
	c := claims{
		Data: adminID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiry)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(s.secret)
}

// Parse validates the token's signature and expiry, returning the admin
// ID embedded in its claims.
func (s *JWTService) Parse(tokenString string) (int64, error) {
	var c claims
	token, err := jwt.ParseWithClaims(tokenString, &c, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return 0, err
	}
	if !token.Valid {
		return 0, errors.New("invalid token")
	}
	return c.Data, nil
}
