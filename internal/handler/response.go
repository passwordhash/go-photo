package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ResponseStatus string

const (
	OkResponse        ResponseStatus = "ok"
	ErrResponse       ResponseStatus = "error"
	PartialOkResponse ResponseStatus = "partial_ok"
)

func NewOkResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

func NewErrResponse(c *gin.Context, code int, respMsg string, err error) {
	outErr := errors.New(respMsg)
	if err != nil {
		outErr = fmt.Errorf("%s: %w", respMsg, err)
	}
	c.Error(outErr)
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
