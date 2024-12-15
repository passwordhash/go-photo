package auth

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go-photo/internal/handler/request"
	"go-photo/internal/handler/response"
	"go-photo/internal/service"
	serviceErr "go-photo/internal/service/error"
	"net/http"
)

type handler struct {
	authService service.UserService
}

func NewHandler(authService service.UserService) *handler {
	return &handler{authService: authService}
}

func (h *handler) RegisterRoutes(router *gin.RouterGroup) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", h.login)
		//authGroup.POST("/register", h.register)
	}
}

// @Summary Login user
// @Description Authenticate a user by email and password, and return a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param input body request.AuthLogin true "Login credentials"
// @Success 200 {object} response.Login
// @Failure 400 {object} response.Error "Invalid request body format."
// @Failure 401 {object} response.Error "Email or password is incorrect."
// @Failure 500 {object} response.Error "Unexpected error occurred."
// @Router /api/v1/auth/login [post]
func (h *handler) login(c *gin.Context) {
	var input request.AuthLogin
	err := c.ShouldBindJSON(&input)
	if err != nil {
		response.NewErr(c, http.StatusBadRequest, response.InvalidRequestParams, err, "Invalid request body format.")
		return
	}

	token, err := h.authService.Login(c, input.Email, input.Password)
	if errors.Is(err, serviceErr.UserNotFoundError) {
		response.NewErr(c, http.StatusUnauthorized, response.InvalidCredentials, err, "Email or password is incorrect.")
		return
	}
	if err != nil {
		response.NewErr(c, http.StatusInternalServerError, response.LoginFailed, err, "Unexpected error occurred.")
		return
	}

	response.NewOk(c, response.Login{Token: token})
}
