package docs

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "go-photo/docs"
)

type handler struct {
}

func NewHandler() *handler {
	return &handler{}
}

func (h *handler) RegisterRoutes(router *gin.RouterGroup) {
	docsGroup := router.Group("/docs")
	{
		docsGroup.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}
