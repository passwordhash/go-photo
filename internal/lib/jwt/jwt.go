package jwt

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserUUID string `json:"uuid"`
	jwt.RegisteredClaims
}

type VerifyTokenFunc func(ctx context.Context, token, secret string) (*Claims, error)

func ValidateToken(_ context.Context, tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
