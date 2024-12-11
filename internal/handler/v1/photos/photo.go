package photos

import (
	"context"
	"errors"
	"fmt"
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
	FormPhotoFile       = "photo_file"
	FormPhotoBatchFiles = "batch_photo_files"
)

func (h *Handler) uploadPhoto(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, config.DefaultContextTimeout)
	defer cancel()

	// TEMP
	UUID := "123e4567-e89b-12d3-a456-426614174000"

	fileHeader, err := c.FormFile(FormPhotoFile)
	if err != nil {
		handler.NewErrResponse(c, http.StatusBadRequest, "file not found", err)
		return
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !utils.IsPhoto(ext) {
		handler.NewErrResponse(c, http.StatusBadRequest, "unsupported file type", nil)
		return
	}

	var alreadyExistsErr *photo.FileAlreadyExistsError
	photoID, err := h.photoService.UploadPhoto(ctx, UUID, fileHeader)
	if errors.As(err, &alreadyExistsErr) {
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

func (h *Handler) uploadBatchPhotos(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, config.DefaultContextTimeout)
	defer cancel()

	// TEMP
	UUID := "123e4567-e89b-12d3-a456-426614174000"

	form, err := c.MultipartForm()
	if err != nil {
		handler.NewErrResponse(c, http.StatusBadRequest, "failed to parse form", err)
		return
	}

	files := form.File[FormPhotoBatchFiles]
	if len(files) == 0 {
		handler.NewErrResponse(c, http.StatusBadRequest, "no batch_photo_files in form", nil)
		return
	}

	if ok, notPhoto := utils.IsAllPhotos(files); !ok {
		handler.NewErrResponse(c, http.StatusBadRequest, "unsupported file type: "+notPhoto, nil)
		return
	}

	photos, err := h.photoService.UploadBatchPhotos(ctx, UUID, files)
	var alreadyExistsErr *photo.FileAlreadyExistsError
	if errors.As(err, &alreadyExistsErr) {
		errMsg := fmt.Sprintf("file with name '%s' already exists in the folder", alreadyExistsErr.Filename)
		handler.NewOkResponse(c, UploadBatchPhotosResponse{
			Status:         handler.PartialOkResponse,
			TotalCount:     len(files),
			SuccessCount:   len(photos),
			UploadedPhotos: append(make([]string, 0), photos...),
			Error:          errMsg,
		})
		return
	}
	if handler.HandleError(c, err) {
		return
	}

	status := handler.OkResponse
	if len(photos) != len(files) {
		status = handler.PartialOkResponse
	}

	handler.NewOkResponse(c, UploadBatchPhotosResponse{
		Status:         status,
		TotalCount:     len(files),
		SuccessCount:   len(photos),
		UploadedPhotos: append(make([]string, 0), photos...),
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
