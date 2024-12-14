package photos

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"go-photo/internal/config"
	"go-photo/internal/handler/response"
	serviceErr "go-photo/internal/service/error"
	"go-photo/internal/service/photo/model"
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

func (h *handler) uploadPhoto(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, config.DefaultContextTimeout)
	defer cancel()

	// TEMP
	UUID := "123e4567-e89b-12d3-a456-426614174000"

	fileHeader, err := c.FormFile(FormPhotoFile)
	if err != nil {
		response.NewErrResponse(c, http.StatusBadRequest, "file not found", err)
		return
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !utils.IsPhoto(ext) {
		response.NewErrResponse(c, http.StatusBadRequest, "unsupported file type", nil)
		return
	}

	var alreadyExistsErr *serviceErr.FileAlreadyExistsError
	photoID, err := h.photoService.UploadPhoto(ctx, UUID, fileHeader)
	if errors.As(err, &alreadyExistsErr) {
		response.NewErrResponse(c, http.StatusBadRequest, "file with the same name already exists", err)
		return
	}
	if response.HandleError(c, err) {
		return
	}

	c.JSON(200, gin.H{
		"status": "ok",
		"id":     photoID,
	})
}

func (h *handler) uploadBatchPhotos(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, config.DefaultContextTimeout)
	defer cancel()

	respStatus := http.StatusOK

	// TEMP
	UUID := "123e4567-e89b-12d3-a456-426614174000"

	form, err := c.MultipartForm()
	if err != nil {
		response.NewErrResponse(c, http.StatusBadRequest, "failed to parse form", err)
		return
	}

	files := form.File[FormPhotoBatchFiles]
	if len(files) == 0 {
		response.NewErrResponse(c, http.StatusBadRequest, "no batch_photo_files in form", nil)
		return
	}

	if ok, notPhoto := utils.IsAllPhotos(files); !ok {
		response.NewErrResponse(c, http.StatusBadRequest, "unsupported file type: "+notPhoto, nil)
		return
	}

	uploads, err := h.photoService.UploadBatchPhotos(ctx, UUID, files)
	if errors.Is(err, serviceErr.AllFailedError) {
		respStatus = http.StatusBadRequest
	} else if errors.Is(err, serviceErr.ParticalSuccessError) {
		respStatus = http.StatusPartialContent
	} else if response.HandleError(c, err) {
		return
	}

	body := response.UploadBatchPhotosResponse{
		TotalCount:   uploads.Total(),
		SuccessCount: uploads.SuccessCount(),
		UploadInfos:  append(make([]response.UploadInfo, 0), model.ToUploadsInfoFromService(uploads.Get())...),
	}

	c.JSON(respStatus, body)
}

func (h *handler) getPhotoVersions(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, config.DefaultContextTimeout)
	defer cancel()

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.NewErrResponse(c, http.StatusBadRequest, "invalid id param", err)
		return
	}

	version, err := h.photoService.GetPhotoVersions(ctx, id)
	if err != nil {
		response.NewErrResponse(c, http.StatusInternalServerError, "failed to get photo versions", err)
		return
	}
	if response.HandleError(c, err) {
		return
	}

	c.JSON(200, gin.H{
		"versions": version,
	})
}
