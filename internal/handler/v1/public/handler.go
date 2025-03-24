package public

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go-photo/internal/handler/response"
	serviceErr "go-photo/internal/service/error"
	"net/http"
)

const (
	publicPhotoParam = "publicToken"
)

func (h *handler) getPublicPhoto(c *gin.Context) {
	tokenParam := c.Param(publicPhotoParam)

	versionQuery := c.Query("version")

	imgData, err := h.photoService.GetPhotoFileByVersionAndToken(c, tokenParam, versionQuery)
	if errors.Is(err, serviceErr.InvalidVersionTypeError) {
		response.NewErr(c, http.StatusBadRequest, response.InvalidReqestsQueryParams, err, "Invalid version type")
		return
	}
	if response.HandleError(c, err) {
		return
	}

	c.Data(200, "image/jpeg", imgData)
}
