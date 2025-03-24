package response

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	serviceErr "go-photo/internal/service/error"
	"net/http"
)

type ErrMessage string

const (
	ErrorStatusExample ErrMessage = "some_error_status"

	NotFound                  ErrMessage = "not_found"
	InternalServerError       ErrMessage = "internal_server_error"
	TimedOut                  ErrMessage = "timed_out"
	InvalidRequestParams      ErrMessage = "invalid_request_params"
	InvalidReqestsQueryParams ErrMessage = "invalid_request_query_params"
	ParamsMissing             ErrMessage = "params_missing"
	UnsupportedFileType       ErrMessage = "unsupported_file_type"
	InvalidCredentials        ErrMessage = "invalid_credentials"
	UserAlreadyExists         ErrMessage = "user_already_exists"
	AuthHeaderEmpty           ErrMessage = "auth_header_empty"
	AuthHeaderInvalid         ErrMessage = "auth_header_invalid"
	AuthTokenInvalid          ErrMessage = "auth_token_invalid"
	Unauthorized              ErrMessage = "unauthorized"
	Forbidden                 ErrMessage = "access_denied"

	PhotoNotFound ErrMessage = "photo_not_found"
)

type Message struct {
	Message string `json:"message"`
}

type Error struct {
	Error   ErrMessage `json:"error"`
	Message string     `json:"message"`
}

func New(c *gin.Context, code int, clientMessage string) {
	c.JSON(code, Message{
		Message: clientMessage,
	})
}

func NewOk(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

func NewErr(c *gin.Context, code int, errMessage ErrMessage, err error, clientMessage string) {
	outErr := errors.New(string(errMessage))
	if err != nil {
		outErr = fmt.Errorf("%s: %w", errMessage, err)
	}
	c.Error(outErr)
	c.AbortWithStatusJSON(code, Error{
		Error:   errMessage,
		Message: clientMessage,
	})
}

func HandleError(c *gin.Context, err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		NewErr(c, http.StatusGatewayTimeout, TimedOut, err, "gateway timeout")
		return true
	}
	if errors.Is(err, serviceErr.AccessDeniedError) {
		NewErr(c, http.StatusForbidden, Forbidden, err, "access denied")
	}
	if errors.Is(err, serviceErr.AccessDeniedError) {
		NewErr(c, http.StatusForbidden, Forbidden, err, "access denied")
	}
	if err != nil {
		NewErr(c, http.StatusInternalServerError, InternalServerError, err, "Unexpected error occurred.")
		return true
	}
	return false
}
