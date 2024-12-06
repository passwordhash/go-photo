package converter

import (
	"go-photo/internal/model"
	repoModel "go-photo/internal/repository/photo/model"
)

func ToPhotoFromRepo(photo *repoModel.Photo, folder *repoModel.Folder, versions []repoModel.PhotoVersion) *model.Photo {
	return &model.Photo{
		ID:         photo.ID,
		UserUUID:   photo.UserUUID,
		Filename:   photo.Filename,
		Folder:     *ToFolderFromRepo(folder),
		Versions:   ToPhotoVersionsFromRepo(versions),
		UploadedAt: photo.UploadedAt.Time,
	}
}

func ToFolderFromRepo(folder *repoModel.Folder) *model.Folder {
	return &model.Folder{
		ID:         folder.ID,
		Folderpath: folder.FolderPath,
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
		ID:          version.PhotoID,
		PhotoID:     version.PhotoID,
		VersionType: model.PhotoVersionType(version.VersionType),
		Filepath:    version.Filepath,
		Width:       version.Width,
		Height:      version.Height,
		Size:        version.Size,
	}
}
