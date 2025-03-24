package photo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"go-photo/internal/model"
	def "go-photo/internal/repository"
	repoErr "go-photo/internal/repository/error"
	repoModel "go-photo/internal/repository/photo/model"
	pkgRepo "go-photo/pkg/repository"
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
		INSERT INTO photos (user_uuid, filename, uploaded_at)
		VALUES ($1, $2, $3)
		RETURNING id`
	err = tx.QueryRowContext(ctx, photosQuery,
		params.UserUUID,
		params.Filename,
		params.SavedAt).Scan(&photoID)
	if err != nil {
		return 0, fmt.Errorf("photo %w: %v", repoErr.InsertError, err)
	}

	photoVersionQuery := `
		INSERT INTO photo_versions (photo_id, filepath, size, height, width, saved_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = tx.ExecContext(ctx,
		photoVersionQuery,
		photoID,
		params.Filepath,
		params.Size,
		params.Height,
		params.Width,
		params.SavedAt)
	if err != nil {
		return 0, fmt.Errorf("version %w: %v", repoErr.InsertError, err)
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	return photoID, nil
}

func (r *repository) CreatePhotoPublishedInfo(ctx context.Context, photoID int) (string, error) {
	query := `
		INSERT INTO published_photo_info (photo_id)
		VALUES ($1)
		RETURNING public_token`

	var publicToken string
	row := r.db.QueryRowContext(ctx, query, photoID)
	err := row.Scan(&publicToken)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == pkgRepo.UniqueViolationErrorCode {
			return "", fmt.Errorf("create photo published info %w: %v", repoErr.ConflictError, err)
		}
		return "", fmt.Errorf("failed to insert published photo info: %w", err)
	}

	return publicToken, nil
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
			return nil, repoErr.NotFoundError
		}
		return nil, err
	}

	return &photo, nil
}

func (r *repository) GetPhotoVersionByToken(ctx context.Context, token string, version model.PhotoVersionType) (*repoModel.PhotoVersion, error) {
	var photoVersion repoModel.PhotoVersion

	query := `
		SELECT pv.id, pv.photo_id, pv.version_type, pv.filepath, pv.size, pv.height, pv.width, pv.saved_at
		FROM published_photo_info ppi
		JOIN photo_versions pv ON ppi.photo_id = pv.photo_id
		WHERE ppi.public_token = $1 AND pv.version_type = $2`

	err := r.db.GetContext(ctx, &photoVersion, query, token, string(version))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repoErr.NotFoundError
		}
		return nil, err
	}

	return &photoVersion, nil
}

func (r *repository) GetPhotoVersions(ctx context.Context, photoID int) ([]repoModel.PhotoVersion, error) {
	var versions []repoModel.PhotoVersion

	query := `
		SELECT id, photo_id, version_type, filepath, size, height, width, saved_at
		FROM photo_versions 
		WHERE photo_id = $1
		ORDER BY size`

	err := r.db.SelectContext(ctx, &versions, query, photoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repoErr.NotFoundError
		}
		return nil, err
	}

	return versions, nil
}

func (r *repository) DeletePhotoPublishedInfo(ctx context.Context, photoID int) error {
	query := `
		DELETE FROM published_photo_info
		WHERE photo_id=$(1)`

	res, err := r.db.ExecContext(ctx, query, photoID)
	if err != nil {
		return fmt.Errorf("failed to delete published photo info: %w", err)
	}

	affectedCnt, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows count: %w", err)
	}
	if affectedCnt < 1 {
		return fmt.Errorf("%w: no rows affected with id %d", repoErr.NotFoundError, photoID)
	}

	return nil
}
