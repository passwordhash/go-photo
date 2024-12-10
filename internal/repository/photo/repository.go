package photo

import (
	"context"
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
		INSERT INTO photos (user_uuid, filename)
		VALUES ($1, $2)
		RETURNING id`
	err = tx.QueryRowContext(ctx, photosQuery, params.UserUUID, params.Filename).Scan(&photoID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert params: %w", err)
	}

	photoVersionQuery := `
		INSERT INTO photo_versions (photo_id, filepath, size)
		VALUES ($1, $2, $3)`
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
