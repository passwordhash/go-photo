package public

import (
	"github.com/gin-gonic/gin"
	"go-photo/internal/service"
)

type handler struct {
	photoService service.PhotoService
}

func NewHandler(photoService service.PhotoService) *handler {
	return &handler{
		photoService: photoService,
	}
}

func (h *handler) RegisterRoutes(router *gin.RouterGroup) {
	publicGroup := router.Group("/p")
	{
		publicGroup.GET("/:publicToken", h.getPublicPhoto)
	}
}
