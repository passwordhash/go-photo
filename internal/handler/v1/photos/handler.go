package photos

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go-photo/internal/service"
	"go-photo/internal/service/photo"
	"go-photo/internal/utils"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	formPhotoFilename = "photoFile"
)

// TEMP
func newErrMessage(c *gin.Context, code int, errMsg string) {
	c.AbortWithStatusJSON(code, gin.H{"message": errMsg})
}

type Handler struct {
	photoService service.PhotoService
}

func NewPhotosHandler(photoService service.PhotoService) *Handler {
	return &Handler{photoService: photoService}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	photosGroup := router.Group("/photos")
	{
		photosGroup.POST("/", h.uploadPhoto)
		photosGroup.GET("/:id", h.getPhotoVersions)
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

func (h *Handler) getPhotoVersions(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		newErrMessage(c, http.StatusBadRequest, "invalid photo id")
		return
	}

	// TODO: модумать насчет контекста
	version, err := h.photoService.GetPhotoVersions(context.TODO(), id)
	// TODO: add error handling
	if err != nil {
		newErrMessage(c, http.StatusInternalServerError, err.Error())
		return
	}

	logrus.Info()
	c.JSON(200, gin.H{
		"versions": version,
	})
}
