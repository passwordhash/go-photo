package photos

import (
	"github.com/gin-gonic/gin"
	"go-photo/internal/handler/middleware"
	"go-photo/internal/service"
)

type handler struct {
	photoService service.PhotoService
	tokenService service.TokenService
}

func NewHandler(photoService service.PhotoService, tokenService service.TokenService) *handler {
	return &handler{
		photoService: photoService,
		tokenService: tokenService,
	}
}

func (h *handler) RegisterRoutes(router *gin.RouterGroup) {
	photosGroup := router.Group("/photos")

	photosGroup.Use(middleware.UserIdentity(h.tokenService.VerifyToken))

	{
		photosGroup.POST("/", h.uploadPhoto)
		photosGroup.POST("/batch", h.uploadBatchPhotos)
		photosGroup.GET("/:id/versions", h.getPhotoVersions)
		photosGroup.GET("/:id/download", h.downloadPhoto)
	}
}
