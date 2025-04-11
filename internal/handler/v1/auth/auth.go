package auth

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go-photo/internal/handler/request"
	"go-photo/internal/handler/response"
	"go-photo/internal/handler/response/auth"
	serviceErr "go-photo/internal/service/error"
	serviceUserModel "go-photo/internal/service/user/model"
	"net/http"
)

// @Summary Login user
// @Description Authenticate a user by email and password, and return a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param input body request.AuthLogin true "Login credentials"
// @Success 200 {object} auth.Login
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
	if response.HandleError(c, err) {
		return
	}

	response.NewOk(c, auth.Login{Token: token})
}

// @Summary Register user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param input body request.AuthRegister true "Registration credentials"
// @Success 200 {object} auth.Register
// @Failure 400 {object} response.Error "Invalid request body format."
// @Failure 409 {object} response.Error "User with this email already exists."
// @Failure 500 {object} response.Error "Unexpected error occurred."
// @Router /api/v1/auth/register [post]
func (h *handler) register(c *gin.Context) {
	var input request.AuthRegister
	err := c.ShouldBindJSON(&input)
	if err != nil {
		response.NewErr(c, http.StatusBadRequest, response.InvalidRequestParams, err, "Invalid request body format.")
		return
	}
	// TODO: validate input

	info, err := h.authService.Register(c, serviceUserModel.RegisterParams{
		Email:    input.Email,
		Password: input.Password,
	})
	if errors.Is(err, serviceErr.UserAlreadyExistsError) {
		response.NewErr(c, http.StatusConflict, response.UserAlreadyExists, err, "User with this email already exists.")
		return
	}
	if response.HandleError(c, err) {
		return
	}

	response.NewOk(c, auth.Register{UserUUID: info.UserUUID, Token: info.Token})
}
