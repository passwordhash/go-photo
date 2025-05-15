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
	tokenServcie service.TokenService,
) *handler {
	return &handler{
		photoService: photoService,
	}
}

func (h *handler) RegisterRoutes(router *gin.RouterGroup) {
	photosGroup := router.Group("/photos")

	photosGroup.Use(middleware.UserIdentity(h.tokenService.ValidateToken))

	{
		photosGroup.POST("/", h.uploadPhoto)
		photosGroup.POST("/batch", h.uploadBatchPhotos)
		{
			photoGroup := photosGroup.Group("/:id")

			photoGroup.GET("/versions", h.getPhotoVersions)
			photoGroup.POST("/publicate", h.publishPhoto)
			photoGroup.DELETE("/unpublicate", h.unpublicatePhoto)
		}

	}
}
