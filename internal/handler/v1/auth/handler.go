package auth

import (
	"go-photo/internal/service"

	"github.com/gin-gonic/gin"
)

type handler struct {
	authService service.AuthService
}

func NewHandler(authService service.AuthService) *handler {
	return &handler{authService: authService}
}

func (h *handler) RegisterRoutes(router *gin.RouterGroup) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", h.register)
		authGroup.POST("/login", h.login)
	}
}
