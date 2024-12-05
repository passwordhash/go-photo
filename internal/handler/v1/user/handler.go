package user

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go-photo/internal/service"
	"go-photo/internal/service/user"
	"go-photo/internal/utils"
	"path/filepath"
	"strconv"
	"strings"
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
		userGroup.POST("/:id/photos", h.uploadPhoto)
	}
}

func (h *Handler) uploadPhoto(c *gin.Context) {
	UUID := "123e4567-e89b-12d3-a456-426614174000"
	file, fileHeader, err := c.Request.FormFile("photoFile")
	if err != nil {
		c.JSON(400, gin.H{"message": "file not provided"})
		return
	}
	defer file.Close()

	fileName := fileHeader.Filename

	ext := strings.ToLower(filepath.Ext(fileName))
	if !utils.IsAllowedExtension(ext) {
		c.JSON(400, gin.H{"message": "invalid file format"})
		return
	}

	size, err := h.userService.UploadFile(c, UUID, file, fileHeader.Filename)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrFailedToCreateFolder):
			c.JSON(500, gin.H{"message": user.ErrFailedToCreateFolder})
		case errors.Is(err, user.ErrFileAlreadyExists):
			c.JSON(409, gin.H{"message": user.ErrFileAlreadyExists})
		case errors.Is(err, user.ErrInvalidFilePermission):
			c.JSON(500, gin.H{"message": user.ErrInvalidFilePermission})
		case errors.Is(err, user.ErrFailedToSavePhoto):
			c.JSON(500, gin.H{"message": user.ErrFailedToSavePhoto})

		}
	}

	c.JSON(200, gin.H{
		"status":     "ok",
		"photo_size": size,
	})
}

func (h *Handler) get(c *gin.Context) {
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
