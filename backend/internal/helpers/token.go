package helpers

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// TempTokenClaims holds claims for temporary 2FA tokens
type TempTokenClaims struct {
	UserID   uint `json:"user_id"`
	TenantID uint `json:"tenant_id"`
	jwt.RegisteredClaims
}

// TempTokenExpiry is the expiration time for temporary 2FA tokens
const TempTokenExpiry = 5 * time.Minute

// GenerateTempToken creates a short-lived token for 2FA verification
func GenerateTempToken(userID, tenantID uint) (string, error) {
	claims := TempTokenClaims{
		UserID:   userID,
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TempTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "2fa_temp",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// ValidateTempToken validates a temporary 2FA token and returns the claims
func ValidateTempToken(tokenString string) (*TempTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TempTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TempTokenClaims); ok && token.Valid {
		// Verify this is a 2FA temp token
		if claims.Subject != "2fa_temp" {
			return nil, fmt.Errorf("invalid token type")
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GenerateToken creates a JWT token for a user (wrapper for auth handler)
func GenerateToken(userID, tenantID uint, role string, isSuperAdmin bool, patientID *uint) (string, error) {
	type Claims struct {
		UserID       uint   `json:"user_id"`
		TenantID     uint   `json:"tenant_id"`
		Role         string `json:"role"`
		IsSuperAdmin bool   `json:"is_super_admin"`
		PatientID    *uint  `json:"patient_id,omitempty"`
		jwt.RegisteredClaims
	}

	claims := Claims{
		UserID:       userID,
		TenantID:     tenantID,
		Role:         role,
		IsSuperAdmin: isSuperAdmin,
		PatientID:    patientID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// GenerateRefreshToken creates a refresh token
func GenerateRefreshToken(userID uint) (string, error) {
	type Claims struct {
		UserID uint `json:"user_id"`
		jwt.RegisteredClaims
	}

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "refresh",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// SetAuthCookies sets the access and refresh token cookies
func SetAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	secure := os.Getenv("GIN_MODE") == "release"
	accessMaxAge := int((15 * time.Minute).Seconds())
	refreshMaxAge := int((7 * 24 * time.Hour).Seconds())

	// Set access token cookie
	c.Header("Set-Cookie", fmt.Sprintf("auth_token=%s; Path=/; Max-Age=%d; HttpOnly; SameSite=Lax%s",
		accessToken, accessMaxAge, func() string {
			if secure {
				return "; Secure"
			}
			return ""
		}()))

	// Set refresh token cookie
	c.Writer.Header().Add("Set-Cookie", fmt.Sprintf("refresh_token=%s; Path=/api/auth; Max-Age=%d; HttpOnly; SameSite=Lax%s",
		refreshToken, refreshMaxAge, func() string {
			if secure {
				return "; Secure"
			}
			return ""
		}()))
}
