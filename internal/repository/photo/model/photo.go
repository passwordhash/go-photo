package model

import (
	"database/sql"
)

type Photo struct {
	ID         int           `db:"id"`
	UserUUID   string        `db:"user_uuid"`
	Filename   string        `db:"filename"`
	UploadedAt *sql.NullTime `db:"uploaded_at"`
}

type PhotoVersion struct {
	ID          int            `db:"id"`
	PhotoID     int            `db:"photo_id"`
	VersionType sql.NullString `db:"version_type"`
	Filepath    string         `db:"filepath"`
	Size        int64          `db:"size"`
}

type CreateOriginalPhotoParams struct {
	UserUUID string
	Filename string
	Filepath string
	Size     int64
}
