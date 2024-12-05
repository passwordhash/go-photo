package photos

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go-photo/internal/service"
	"go-photo/internal/service/photo"
	"go-photo/internal/utils"
	"net/http"
	"path/filepath"
	"strings"
)

const (
	formPhotoFilename = "photoFile"
)

// TEMP
func newErrMessage(ctx *gin.Context, code int, errMsg string) {
	ctx.AbortWithStatusJSON(code, gin.H{"message": errMsg})
}

type Handler struct {
	photoService service.PhotoService
}

func NewPhotosHandler(photoService service.PhotoService) *Handler {
	return &Handler{photoService: photoService}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	userGroup := router.Group("/photos")
	{
		userGroup.POST("/", h.uploadPhoto)
	}
}

func (h *Handler) uploadPhoto(c *gin.Context) {
	// TEMP
	UUID := "123e4567-e89b-12d3-a456-426614174000"

	file, fileHeader, err := c.Request.FormFile(formPhotoFilename)
	if err != nil {
		newErrMessage(c, http.StatusBadRequest, "file not provided")
		return
	}
	defer file.Close()

	fileName := fileHeader.Filename

	ext := strings.ToLower(filepath.Ext(fileName))
	if !utils.IsPhoto(ext) {
		newErrMessage(c, http.StatusBadRequest, "invalid file format")
		return
	}

	size, err := h.photoService.UploadPhoto(c, UUID, file, fileHeader.Filename)
	if errors.Is(err, photo.FileAlreadyExistsError) {
		newErrMessage(c, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		newErrMessage(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(200, gin.H{
		"status":     "ok",
		"photo_size": size,
	})
}
