package photos

import "go-photo/internal/handler"

type UploadBatchPhotosResponse struct {
	Status         handler.ResponseStatus `json:"status"`
	TotalCount     int                    `json:"total_count"`
	SuccessCount   int                    `json:"success_count"`
	UploadedPhotos []string               `json:"uploaded_photos"`
}
