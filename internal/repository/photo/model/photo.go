package model

import (
	"database/sql"
)

type Photo struct {
	ID         int           `db:"id"`
	UserUUID   string        `db:"user_uuid"`
	Filename   string        `db:"filename"`
	FolderID   int           `db:"folder_id"`
	UploadedAt *sql.NullTime `db:"uploaded_at"`
}

type PhotoVersion struct {
	ID          int            `db:"id"`
	PhotoID     int            `db:"photo_id"`
	VersionType sql.NullString `db:"version_type"`
	Filepath    string         `db:"filepath"`
	//Width       int            `db:"width"`
	//Height      int            `db:"height"`
	Size int64 `db:"size"`
}

type Folder struct {
	//ID         int    `db:"id"`
	FolderPath string `db:"folder_path"`
	UserUUID   string `db:"user_uuid"`
}
