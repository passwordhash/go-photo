package photos

import (
	"github.com/gin-gonic/gin"
	"go-photo/internal/service"
)

type handler struct {
	photoService service.PhotoService
}

func NewHandler(photoService service.PhotoService) *handler {
	return &handler{photoService: photoService}
}

func (h *handler) RegisterRoutes(router *gin.RouterGroup) {
	photosGroup := router.Group("/photos")
	{
		photosGroup.POST("/", h.uploadPhoto)
		photosGroup.POST("/batch", h.uploadBatchPhotos)
		photosGroup.GET("/:id", h.getPhotoVersions)
	}
}
