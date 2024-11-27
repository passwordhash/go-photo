package user

import (
	"github.com/gin-gonic/gin"
	"go-photo/internal/service"
)

type Handler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *Handler {
	return &Handler{userService: userService}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	userGroup := router.Group("/user")
	{
		userGroup.GET("/:id", h.Get)
	}
}

func (h *Handler) Get(c *gin.Context) {
	userId := c.Param("id")

	c.JSON(200, gin.H{
		"userId": userId,
	})
}
