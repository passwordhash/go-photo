package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"go-photo/internal/handler/response"
	serviceUserModel "go-photo/internal/service/user/model"
	"net/http"
	"strings"
)

const (
	authorizationHeader = "Authorization"
	UserUUIDCtx         = "user"
)

type VerifyTokenFunc func(ctx context.Context, token string) (serviceUserModel.TokenPayload, error)

func UserIdentity(verifyFucn VerifyTokenFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIdentity(c, verifyFucn)
		c.Next()
	}
}

func userIdentity(c *gin.Context, verify VerifyTokenFunc) {
	header := c.GetHeader(authorizationHeader)

	if header == "" {
		response.NewErr(c, http.StatusUnauthorized, response.AuthHeaderEmpty, nil, "Auth header is empty.")
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		response.NewErr(c, http.StatusUnauthorized, response.AuthHeaderInvalid, nil, "Bearer token is invalid.")
		return
	}

	token := headerParts[1]
	if token == "" {
		response.NewErr(c, http.StatusUnauthorized, response.AuthHeaderInvalid, nil, "Token is empty.")
		return
	}

	resp, err := verify(context.Background(), token)
	if err != nil {
		response.NewErr(c, http.StatusUnauthorized, response.AuthTokenInvalid, err, "Token is invalid or user cannot be found.")
		return
	}

	c.Set(UserUUIDCtx, resp.UserUUID)
}
