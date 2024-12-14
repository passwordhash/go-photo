package auth

import (
	"github.com/gin-gonic/gin"
	"go-photo/internal/handler/request"
	"go-photo/internal/handler/response"
	"go-photo/internal/service"
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

func (h *handler) login(c *gin.Context) {
	var input request.AuthLogin
	err := c.ShouldBindJSON(&input)
	if err != nil {
		response.NewErrResponse(c, 400, "invalid request body", err)
		return
	}

	token, err := h.authService.Login(c, input.Email, input.Password)
	if err != nil {
		response.NewErrResponse(c, 401, "login failed", err)
		return
	}

	response.NewOkResponse(c, gin.H{"token": token})
}
