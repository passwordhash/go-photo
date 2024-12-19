package photo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	def "go-photo/internal/repository"
	repoErr "go-photo/internal/repository/error"
	repoModel "go-photo/internal/repository/photo/model"
	"os"
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

func (r *repository) DeletePhoto(ctx context.Context, photoID int) (error, error) {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", repoErr.BeginTxError, err), nil
	}

	defer func() {
		if err != nil {
			err = tx.Rollback()
			log.Errorf("failed to rollback transaction: %v\ncontext: %v", err, ctx)
		}
	}()

	pathQuery := `
		SELECT filepath	FROM photo_versions
		WHERE photo_id = $1`
	rows, err := tx.QueryContext(ctx, pathQuery, photoID)
	if err != nil {
		return fmt.Errorf("failed to get photo path: %w", err), nil
	}
	defer rows.Close()

	var filepaths []string
	for rows.Next() {
		var filepath string
		if err := rows.Scan(&filepath); err != nil {
			log.Fatalf("failed to scan photo path: %v", err)
		}
		filepaths = append(filepaths, filepath)
	}

	versionsDeleteQuery := `
		DELETE FROM photo_versions
		WHERE photo_id = $1`
	_, err = tx.ExecContext(ctx, versionsDeleteQuery, photoID)
	if err != nil {
		return fmt.Errorf("failed to delete photo versions: %w", err), nil
	}

	photosDeleteQuery := `
		DELETE FROM photos
		WHERE id = $1`
	_, err = tx.ExecContext(ctx, photosDeleteQuery, photoID)
	if err != nil {
		return fmt.Errorf("failed to delete photo: %w", err), nil
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: %v", repoErr.CommitTxError, err), nil
	}

	for _, filepath := range filepaths {
		if err := os.Remove(filepath); err != nil {
			log.Printf("failed to remove file %s: %v", filepath, err)
		}
	}

	return nil, nil
}

func (r *repository) CreateOriginalPhoto(ctx context.Context, params *repoModel.CreateOriginalPhotoParams) (int, error) {
	if params == nil {
		return 0, repoErr.NilParamsError
	}
	if !params.IsValid() {
		return 0, fmt.Errorf("%w: %v", repoErr.InvalidParamsError, params)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", repoErr.BeginTxError, err)
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
		return 0, fmt.Errorf("photo %w: %v", repoErr.InsertError, err)
	}

	photoVersionQuery := `
		INSERT INTO photo_versions (photo_id, filepath, size)
		VALUES ($1, $2, $3)`
	_, err = tx.ExecContext(ctx, photoVersionQuery, photoID, params.Filepath, params.Size)
	if err != nil {
		return 0, fmt.Errorf("version %w: %v", repoErr.InsertError, err)
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
			return nil, repoErr.PhotoNotFound
		}
		return nil, err
	}

	return &photo, nil
}

func (r *repository) GetPhotoVersions(ctx context.Context, photoID int) ([]repoModel.PhotoVersion, error) {
	var versions []repoModel.PhotoVersion

	_, err := r.GetPhotoByID(ctx, photoID)
	if errors.Is(err, repoErr.PhotoNotFound) {
		return nil, repoErr.PhotoNotFound
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
