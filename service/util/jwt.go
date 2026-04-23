package util

import (
	"SService/config"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// 自定义claims
type Claims struct {
	ID       int       `json:"id"`
	Email    string    `json:"email"`
	Nickname string    `json:"nickname"`
	UUID     uuid.UUID `json:"uuid"`
	jwt.RegisteredClaims
}

// 生成JWT令牌
func GenerateToken(userID int, email, nickname string, userUUID uuid.UUID) (string, error) {
	cfg := config.AppConfig.JWT
	expireHours := cfg.ExpireHours
	if expireHours <= 0 {
		expireHours = 240
	}
	issuer := cfg.Issuer
	if issuer == "" {
		issuer = "SService"
	}
	secret := cfg.Secret
	if secret == "" {
		return "", errors.New("jwt secret 未配置")
	}

	expirationTime := time.Now().Add(time.Duration(expireHours) * time.Hour)

	claims := &Claims{
		ID:       userID,
		Email:    email,
		Nickname: nickname,
		UUID:     userUUID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
		},
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// 解析JWT令牌
func ParseToken(tokenString string) (*Claims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		secret := config.AppConfig.JWT.Secret
		if secret == "" {
			return nil, errors.New("jwt secret 未配置")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证token并提取claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func GetUserInfo(c *gin.Context) *Claims {
	claims, exists := c.Get("claims")
	if !exists {
		return nil
	}
	return claims.(*Claims)
}
