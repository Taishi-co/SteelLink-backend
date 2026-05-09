package auth

import (
	"errors"
	"time"

	"github.com/Taishi-co/SteelLink-backend/config"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID              uint64   `json:"user_id"`
	Phone               string   `json:"phone,omitempty"`
	Email               string   `json:"email,omitempty"`
	UserType            string   `json:"user_type"`
	EnterpriseID        uint64   `json:"enterprise_id,omitempty"`
	EnterpriseCode      string   `json:"enterprise_code,omitempty"`
	EnterpriseRole      string   `json:"enterprise_role,omitempty"`
	ServiceCapabilities []string `json:"service_capabilities,omitempty"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

var (
	ErrTokenInvalid = errors.New("token is invalid")
	ErrTokenExpired = errors.New("token is expired")
)

// GenerateToken generates access token and refresh token
func GenerateToken(userID uint64, phone, email, userType string, enterpriseID uint64, enterpriseCode, enterpriseRole string, serviceCapabilities ...[]string) (*TokenPair, error) {
	cfg := config.Get()

	now := time.Now()
	expiresAt := now.Add(cfg.JWT.ExpireDuration())
	refreshExpiresAt := now.Add(cfg.JWT.RefreshExpireDuration())

	var capabilities []string
	if len(serviceCapabilities) > 0 {
		capabilities = serviceCapabilities[0]
	}

	// Access token claims
	accessClaims := Claims{
		UserID:              userID,
		Phone:               phone,
		Email:               email,
		UserType:            userType,
		EnterpriseID:        enterpriseID,
		EnterpriseCode:      enterpriseCode,
		EnterpriseRole:      enterpriseRole,
		ServiceCapabilities: capabilities,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.JWT.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		return nil, err
	}

	// Refresh token claims
	refreshClaims := Claims{
		UserID:              userID,
		UserType:            userType,
		EnterpriseID:        enterpriseID,
		ServiceCapabilities: capabilities,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.JWT.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(cfg.JWT.Expire),
	}, nil
}

// ParseToken parses and validates the access token
func ParseToken(tokenString string) (*Claims, error) {
	cfg := config.Get()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ecode.ErrTokenExpired
		}
		return nil, ecode.ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ecode.ErrTokenInvalid
}

// RefreshToken generates new token pair from refresh token
func RefreshToken(refreshTokenString string) (*TokenPair, error) {
	claims, err := ParseToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	return GenerateToken(claims.UserID, claims.Phone, claims.Email, claims.UserType, claims.EnterpriseID, claims.EnterpriseCode, claims.EnterpriseRole, claims.ServiceCapabilities)
}
