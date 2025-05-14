package middleware

import (
	"go-photo/internal/handler/response"
	"go-photo/internal/lib/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	authorizationHeader = "Authorization"
	UserUUIDKey         = "user"
)

func UserIdentity(verifyFuc jwt.VerifyTokenFunc, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIdentity(c, verifyFuc, secret)
		c.Next()
	}
}

func userIdentity(c *gin.Context, verifyFuc jwt.VerifyTokenFunc, secret string) {
	header := c.GetHeader(authorizationHeader)

	if header == "" {
		response.NewErr(c,
			http.StatusUnauthorized,
			response.AuthHeaderEmpty,
			nil, "Auth header is empty.",
		)
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		response.NewErr(c,
			http.StatusUnauthorized,
			response.AuthHeaderInvalid,
			nil, "Bearer token is invalid.",
		)
		return
	}

	token := headerParts[1]
	if token == "" {
		response.NewErr(c,
			http.StatusUnauthorized,
			response.AuthHeaderInvalid,
			nil, "Token is empty.",
		)
		return
	}

	// claims, err := jwt.ValidateToken(token, secret)
	claims, err := verifyFuc(c, token, secret)
	if err != nil {
		response.NewErr(c,
			http.StatusUnauthorized,
			response.AuthTokenInvalid,
			err, "Token is invalid or user cannot be found.",
		)
		return
	}

	c.Set(UserUUIDKey, claims.UserUUID)
}
