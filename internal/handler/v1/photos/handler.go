package photos

import (
	"go-photo/internal/handler/middleware"
	"go-photo/internal/service"

	"github.com/gin-gonic/gin"
)

type handler struct {
	photoService service.PhotoService
	tokenService service.TokenService
}

func NewHandler(
	photoService service.PhotoService,
	tokenService service.TokenService,
) *handler {
	return &handler{
		photoService: photoService,
		tokenService: tokenService,
	}
}

func (h *handler) RegisterRoutes(router *gin.RouterGroup) {
	photosGroup := router.Group("/photos")

	{
		photosGroup.Use(middleware.UserIdentity(h.tokenService.ValidateToken))

		photosGroup.POST("/", h.uploadPhoto)

		photosGroup.POST("/batch", h.uploadBatchPhotos)

		{
			photoGroup := photosGroup.Group("/:id")

			photoGroup.Use(middleware.PhotoOwner(h.photoService))

			photoGroup.GET("/versions", h.getPhotoVersions)
			photoGroup.POST("/publish", h.publishPhoto)
			photoGroup.DELETE("/unpublish", h.unpublishPhoto)
		}
	}
}
