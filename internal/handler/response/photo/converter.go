package photo

import (
	"go-photo/internal/model"
	"time"
)

func ToPhotoVersionFromModel(photoVersion model.PhotoVersion) PhotoVersion {
	return PhotoVersion{
		PhotoID:     photoVersion.PhotoID,
		VersionType: string(photoVersion.VersionType),
		Filepath:    photoVersion.UUIDFilename,
		Size:        photoVersion.Size,
		Height:      photoVersion.Height,
		Width:       photoVersion.Width,
		SavedAt:     photoVersion.SavedAt.Format(time.DateTime),
	}
}

func ToPhotoVersionsFromModel(photoVersions []model.PhotoVersion) []PhotoVersion {
	photoVersionsResponse := make([]PhotoVersion, len(photoVersions))
	for i, pv := range photoVersions {
		photoVersionsResponse[i] = ToPhotoVersionFromModel(pv)
	}
	return photoVersionsResponse
}
