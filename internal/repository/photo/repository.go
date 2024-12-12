package photo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
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
	if params == nil {
		return 0, def.NilParamsError
	}
	if !params.IsValid() {
		return 0, fmt.Errorf("%w: %v", def.InvalidParamsError, params)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", def.BeginTxError, err)
	}

	defer func() {
		if err != nil {
			err = tx.Rollback()
			log.Errorf("failed to rollback transaction: %v\ncontext: %v", err, ctx)
		}
	}()

	var photoID int
	photosQuery := `
		INSERT INTO photos (user_uuid, filename)
		VALUES ($1, $2)
		RETURNING id`
	err = tx.QueryRowContext(ctx, photosQuery, params.UserUUID, params.Filename).Scan(&photoID)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", def.InsertPhotoError, err)
	}

	photoVersionQuery := `
		INSERT INTO photo_versions (photo_id, filepath, size)
		VALUES ($1, $2, $3)`
	_, err = tx.ExecContext(ctx, photoVersionQuery, photoID, params.Filepath, params.Size)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", def.InsertVersionError, err)
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	return photoID, nil
}

func (r *repository) GetPhotoByID(ctx context.Context, photoID int) (*repoModel.Photo, error) {
	var photo repoModel.Photo

	query := `
		SELECT id, user_uuid, filename, uploaded_at
		FROM photos
		WHERE id = $1`

	err := r.db.GetContext(ctx, &photo, query, photoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, def.PhotoNotFound
		}
		return nil, err
	}

	return &photo, nil
}

func (r *repository) GetPhotoVersions(ctx context.Context, photoID int) ([]repoModel.PhotoVersion, error) {
	var versions []repoModel.PhotoVersion

	_, err := r.GetPhotoByID(ctx, photoID)
	if errors.Is(err, def.PhotoNotFound) {
		return nil, def.PhotoNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get photo by id: %w", err)
	}

	query := `
		SELECT id, photo_id, version_type, filepath, size 
		FROM photo_versions 
		WHERE photo_id = $1
		ORDER BY size`

	err = r.db.Select(&versions, query, photoID)
	if err != nil {
		return nil, err
	}

	return versions, nil
}
