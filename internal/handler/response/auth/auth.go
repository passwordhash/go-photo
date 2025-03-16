package auth

import (
	"github.com/gin-gonic/gin"
	"go-photo/internal/handler/response"
	"net/http"
)

type Login struct {
	Token string `json:"token"`
}

type Register struct {
	UserUUID string `json:"user_uuid"`
	Token    string `json:"token"`
}

func MustGetUUID(c *gin.Context, key string) (string, bool) {
	val, exists := c.Get(key)
	if !exists {
		response.NewErr(c, http.StatusUnauthorized, response.Unauthorized, nil, "Try logging in again.")
		return "", false
	}

	uuid, ok := val.(string)
	if !ok {
		response.NewErr(c, http.StatusInternalServerError, response.InternalServerError, nil, "Unexpected error occurred.")
		return "", false
	}

	return uuid, true
}
