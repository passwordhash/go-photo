package user

import (
	"github.com/gin-gonic/gin"
	"go-photo/internal/service"
	"strconv"
)

type handler struct {
	userService service.UserService
}

func NewHandler(userService service.UserService) *handler {
	return &handler{userService: userService}
}

func (h *handler) RegisterRoutes(router *gin.RouterGroup) {
	userGroup := router.Group("/users")
	{
		userGroup.GET("/:id", h.get)
		userGroup.GET("/", h.getAll)
	}
}

func (h *handler) get(c *gin.Context) {
	userId := c.Param("id")

	// Пример валидации параметра
	_, err := strconv.Atoi(userId)
	if err != nil {
		c.Error(err)
		c.JSON(400, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.userService.Get(c, userId)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, user)
}

func (h *handler) getAll(c *gin.Context) {
	users, err := h.userService.GetAll(c)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, users)
}
