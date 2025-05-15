package token

import (
	"context"
	"fmt"

	def "go-photo/internal/service"
	serviceTokenModel "go-photo/internal/service/token/model"

	"github.com/golang-jwt/jwt/v5"
)

var _ def.TokenService = (*service)(nil)

type service struct {
	appSecret string
}

func New(appSecret string) *service {
	return &service{
		appSecret: appSecret,
	}
}

type VerifyTokenFunc func(ctx context.Context, token string) (*serviceTokenModel.Claims, error)

func (s *service) ValidateToken(_ context.Context, tokenString string) (*serviceTokenModel.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &serviceTokenModel.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.appSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*serviceTokenModel.Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
