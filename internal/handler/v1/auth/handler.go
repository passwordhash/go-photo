package auth

import (
	"github.com/gin-gonic/gin"
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
