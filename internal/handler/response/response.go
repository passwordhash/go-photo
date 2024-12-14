package response

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ErrMessage string

const (
	InternalServerError  ErrMessage = "internal_server_error"
	TimedOut                        = "timed_out"
	InvalidRequestParams            = "invalid_request_params"
	ParamsMissing                   = "params_missing"
	UnsupportedFileType             = "unsupported_file_type"
	InvalidCredentials              = "invalid_credentials"
	LoginFailed                     = "login_failed"
)

type ErrorResponse struct {
	Error   ErrMessage `json:"error"`
	Message string     `json:"message"`
}

func NewOk(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

func NewErr(c *gin.Context, code int, respMsg ErrMessage, err error, clientMessage string) {
	outErr := errors.New(string(respMsg))
	if err != nil {
		outErr = fmt.Errorf("%s: %w", respMsg, err)
	}
	c.Error(outErr)
	c.AbortWithStatusJSON(code, ErrorResponse{
		Error:   respMsg,
		Message: clientMessage,
	})
}

func HandleError(c *gin.Context, err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		NewErr(c, http.StatusGatewayTimeout, TimedOut, err, "gateway timeout")
		return true
	}
	if err != nil {
		NewErr(c, http.StatusInternalServerError, InternalServerError, err, "internal server error")
		return true
	}
	return false
}
