package middleware

import (
	"go-photo/internal/handler/response"
	"go-photo/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	photoIDParamKey = "id"
)

func PhotoOwner(photoService service.PhotoService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userUUID, exists := c.Get(UserUUIDKey)
		if !exists || userUUID == "" {
			response.NewErr(c,
				http.StatusUnauthorized,
				response.Unauthorized,
				nil, "User not authorized.")
			return
		}

		photoID, err := strconv.Atoi(c.Param(photoIDParamKey))
		if err != nil {
			response.NewErr(c,
				http.StatusBadRequest,
				response.InvalidReqestQueryParams,
				err, "Photo ID is invalid.")
			return
		}

		photo, err := photoService.PhotoByID(c, photoID)
		if err != nil {
			response.NewErr(c,
				http.StatusNotFound,
				response.PhotoNotFound,
				err, "Photo not found.")
			return
		}

		if photo.UserUUID != userUUID.(string) {
			response.NewErr(c,
				http.StatusForbidden,
				response.PhotoNotOwner,
				err, "You are not the owner of this photo.")
			return
		}

		c.Next()
	}
}
