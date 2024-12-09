package photo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	def "go-photo/internal/repository"
	repoModel "go-photo/internal/repository/photo/model"
)

var _ def.PhotoRepository = (*repository)(nil)

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *repository {
	return &repository{
		db: db,
	}
}

func (r *repository) GetFolderID(ctx context.Context, folderpath, userUUID string) (int, error) {
	var folderID int

	query := `
		SELECT id
		FROM Folders
		WHERE folder_path = $1 AND user_uuid = $2
	`

	err := r.db.GetContext(ctx, &folderID, query, folderpath, userUUID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, def.EmptyResultError
	}
	if err != nil {
		return 0, fmt.Errorf("cannot get folder: %w", err)
	}

	return folderID, nil
}

func (r *repository) CreateFolder(ctx context.Context, folderpath, userUUID string) (int, error) {
	var folderID int

	query := `
		INSERT INTO folders (folder_path, user_uuid)
		VALUES ($1, $2)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query, folderpath, userUUID).Scan(&folderID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert folder: %w", err)
	}

	return folderID, nil
}

func (r *repository) MustGetFolder(ctx context.Context, folderpath, userUUID string) (int, error) {
	folderID, err := r.GetFolderID(ctx, folderpath, userUUID)
	if err == nil {
		return folderID, nil
	}
	if !errors.Is(err, def.EmptyResultError) {
		return 0, fmt.Errorf("failed to get folder: %w", err)
	}

	folderID, err = r.CreateFolder(ctx, folderpath, userUUID)
	if err != nil {
		return 0, fmt.Errorf("failed to create folder: %w", err)
	}

	return folderID, nil
}

func (r *repository) CreateOriginalPhoto(ctx context.Context, params *repoModel.CreateOriginalPhotoParams) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var photoID int
	photosQuery := `
		INSERT INTO photos (user_uuid, filename, folder_id)
		VALUES ($1, $2, $3)
		RETURNING id`
	err = tx.QueryRowContext(ctx, photosQuery, params.UserUUID, params.Filename, params.FolderID).Scan(&photoID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert params: %w", err)
	}

	photoVersionQuery := `
		INSERT INTO photo_versions (photo_id, filepath, size)
		VALUES ($1, $2, $3)`

	// TEMP
	_, err = tx.ExecContext(ctx, photoVersionQuery, photoID, params.Filepath, params.Size)
	if err != nil {
		return 0, fmt.Errorf("failed to insert params version: %w", err)
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	return photoID, nil
}

func (r *repository) GetPhotoVersions(_ context.Context, photoID int) ([]repoModel.PhotoVersion, error) {
	var versions []repoModel.PhotoVersion

	query := `
		SELECT id, photo_id, version_type, filepath, size 
		FROM photo_versions 
		WHERE photo_id = $1`

	err := r.db.Select(&versions, query, photoID)
	if err != nil {
		return nil, err
	}

	return versions, err
}
