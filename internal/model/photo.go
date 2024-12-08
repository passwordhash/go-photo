package model

import "time"

type PhotoVersionType string

const (
	Original  PhotoVersionType = "original"
	Thumbnail                  = "thumbnail"
	Preview                    = "preview"
)

type Photo struct {
	ID       int
	UserUUID string
	Filename string
	Folder
	Versions   []PhotoVersion
	UploadedAt time.Time
}

type PhotoVersion struct {
	ID          int
	PhotoID     int
	VersionType PhotoVersionType
	Filepath    string
	//Width       int
	//Height      int
	Size int64
}

type Folder struct {
	ID         int
	Folderpath string
}
