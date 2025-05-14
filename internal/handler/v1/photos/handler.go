package photos

import (
	"go-photo/internal/handler/middleware"
	"go-photo/internal/lib/jwt"
	"go-photo/internal/service"

	"github.com/gin-gonic/gin"
)

type handler struct {
	photoService service.PhotoService

	appSecret string
}

func NewHandler(photoService service.PhotoService, appSecret string) *handler {
	return &handler{
		photoService: photoService,
		appSecret:    appSecret,
	}
}

func (h *handler) RegisterRoutes(router *gin.RouterGroup) {
	photosGroup := router.Group("/photos")

	photosGroup.Use(middleware.UserIdentity(jwt.ValidateToken, h.appSecret))

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
