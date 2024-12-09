package photos

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"go-photo/internal/config"
	"go-photo/internal/handler"
	"go-photo/internal/service/photo"
	"go-photo/internal/utils"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	formPhotoFilename = "photoFile"
)

func (h *Handler) uploadPhoto(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, config.DefaultContextTimeout)
	defer cancel()

	// TEMP
	UUID := "123e4567-e89b-12d3-a456-426614174000"

	file, fileHeader, err := c.Request.FormFile(formPhotoFilename)
	if err != nil {
		handler.NewErrResponse(c, http.StatusBadRequest, "file not found", err)
		return
	}
	defer file.Close()

	fileName := fileHeader.Filename

	ext := strings.ToLower(filepath.Ext(fileName))
	if !utils.IsPhoto(ext) {
		handler.NewErrResponse(c, http.StatusBadRequest, "file is not a photo", nil)
		return
	}

	photoID, err := h.photoService.UploadPhoto(ctx, UUID, file, fileHeader.Filename)
	if errors.Is(err, photo.FileAlreadyExistsError) {
		handler.NewErrResponse(c, http.StatusBadRequest, "file with the same name already exists", err)
		return
	}
	if handler.HandleError(c, err) {
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"id":     photoID,
	})
}

func (h *Handler) getPhotoVersions(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, config.DefaultContextTimeout)
	defer cancel()

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		handler.NewErrResponse(c, http.StatusBadRequest, "invalid id param", err)
		return
	}

	version, err := h.photoService.GetPhotoVersions(ctx, id)
	if err != nil {
		handler.NewErrResponse(c, http.StatusInternalServerError, "failed to get photo versions", err)
		return
	}
	if handler.HandleError(c, err) {
		return
	}

	c.JSON(200, gin.H{
		"versions": version,
	})
}
