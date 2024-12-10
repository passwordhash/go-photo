package photos

import (
	"github.com/gin-gonic/gin"
	"go-photo/internal/service"
)

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
		photosGroup.POST("/batch", h.uploadBatchPhotos)
		photosGroup.GET("/:id", h.getPhotoVersions)
	}
}
