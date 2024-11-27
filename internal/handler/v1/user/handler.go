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

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	userGroup := router.Group("/users")
	{
		userGroup.GET("/:id", h.get)
	}
}

func (h *Handler) get(c *gin.Context) {
	userId := c.Param("id")

	user, err := h.userService.Get(c, userId)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, user)
}
