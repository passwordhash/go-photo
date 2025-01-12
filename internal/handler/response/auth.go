package response

import (
	"github.com/gin-gonic/gin"
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
		NewErr(c, http.StatusUnauthorized, Unauthorized, nil, "Try logging in again.")
		return "", false
	}

	uuid, ok := val.(string)
	if !ok {
		NewErr(c, http.StatusInternalServerError, InternalServerError, nil, "Unexpected error occurred.")
		return "", false
	}

	return uuid, true
}
