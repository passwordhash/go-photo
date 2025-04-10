package converter

import (
	"go-photo/internal/model"
	repoModel "go-photo/internal/repository/photo/model"
)

func ToPhotoFromRepo(photo *repoModel.Photo, versions []repoModel.PhotoVersion) *model.Photo {
	return &model.Photo{
		ID:         photo.ID,
		UserUUID:   photo.UserUUID,
		Filename:   photo.Filename,
		Versions:   ToPhotoVersionsFromRepo(versions),
		UploadedAt: photo.UploadedAt.Time,
	}
}

func ToPhotoVersionsFromRepo(versions []repoModel.PhotoVersion) []model.PhotoVersion {
	var res []model.PhotoVersion

	for _, v := range versions {
		res = append(res, ToPhotoVersionFromRepo(v))
	}

	return res
}

func ToPhotoVersionFromRepo(version repoModel.PhotoVersion) model.PhotoVersion {
	return model.PhotoVersion{
		ID:           version.PhotoID,
		PhotoID:      version.PhotoID,
		VersionType:  model.PhotoVersionType(version.VersionType.String),
		UUIDFilename: version.UUIDFilename,
		Size:         version.Size,
		Height:       version.Height,
		Width:        version.Width,
		SavedAt:      version.SavedAt.Time,
	}
}
