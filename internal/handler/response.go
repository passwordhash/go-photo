package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func NewErrResponse(c *gin.Context, code int, respMsg string, err error) {
	c.Error(fmt.Errorf("%s: %w", respMsg, err))
	c.AbortWithStatusJSON(code, gin.H{"message": respMsg})
}

func HandleError(c *gin.Context, err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		NewErrResponse(c, http.StatusGatewayTimeout, "operation timed out, please try again later", err)
		return true
	}
	if err != nil {
		NewErrResponse(c, http.StatusInternalServerError, "internal server error", err)
		return true
	}
	return false
}
