package middlewares

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	Secret string
	Issuer string
}

func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{
		Secret: secret,
		Issuer: "fly-box-api",
	}
}

func (j *JWTManager) Generate(userID uint, email, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    j.Issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.Secret))
}

func (j *JWTManager) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 1. Kiểm tra Signing Method (Tránh tấn công đổi thuật toán)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return []byte(j.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	// 2. Kiểm tra token có hợp lệ không
	if !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// GetClaims helper để lấy thông tin user từ context
func GetClaims(c *gin.Context) *Claims {
	claimsAny, exists := c.Get("claims")
	if !exists {
		return nil
	}
	claims, ok := claimsAny.(*Claims)
	if !ok {
		return nil
	}
	return claims
}

func AuthRequired(jwtMgr *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}
		tokenString := parts[1]

		// Parse and validate JWT
		claims, err := jwtMgr.Parse(tokenString)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("claims", claims)
		c.Next()
	}
}
