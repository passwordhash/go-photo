package auth

import (
	"go-photo/internal/handler/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Login struct {
	Token string `json:"token"`
}

type Register struct {
	UserUUID string `json:"user_uuid"`
}

func MustGetUUID(c *gin.Context, key string) (string, bool) {
	val, exists := c.Get(key)
	if !exists {
		response.NewErr(c, http.StatusUnauthorized, response.Unauthorized, nil, "Try logging in again.")
		return "", false
	}

	userUUID, ok := val.(string)
	if !ok {
		response.NewErr(c, http.StatusInternalServerError, response.InternalServerError, nil, "Unexpected error occurred.")
		return "", false
	}

	if strings.TrimSpace(userUUID) == "" {
		response.NewErr(c, http.StatusUnauthorized, response.Unauthorized, nil, "Try logging in again.")
		return "", false
	}

	return userUUID, true
}
