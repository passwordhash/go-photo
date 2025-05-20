package photo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	domainModel "go-photo/internal/model"
	repoErr "go-photo/internal/repository/error"
	"go-photo/internal/repository/photo/model"
	pkgRepo "go-photo/pkg/repository"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	photoColumns        = []string{"id", "user_uuid", "filename", "uploaded_at"}
	photoVersionColumns = []string{"id", "photo_id", "version_type", "uuid_filename", "size", "height", "width", "saved_at"}
	idColumn            = []string{"id"}
)

func TestRepository_CreateOriginalPhoto(t *testing.T) {
	repoErraultParams := model.CreateOriginalPhotoParams{
		UserUUID:     "user-uuid",
		Filename:     "test.png",
		UUIDFilename: "home/user-uuid/test.png",
		Size:         12345,
		Height:       100,
		Width:        100,
		SavedAt:      time.Now(),
	}

	tests := []struct {
		name          string
		params        *model.CreateOriginalPhotoParams
		mockSetup     func(mock sqlmock.Sqlmock)
		expectedID    int
		expectedError error
	}{
		{
			name:   "Valid",
			params: &repoErraultParams,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectQuery("INSERT INTO photos").
					WithArgs("user-uuid", "test.png", sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows(idColumn).
						AddRow(1))

				mock.ExpectExec("INSERT INTO photo_versions").
					WithArgs(1, "home/user-uuid/test.png", 12345, 100, 100, sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
			expectedID:    1,
			expectedError: nil,
		},
		{
			name:   "Failed begin transaction",
			params: &repoErraultParams,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(repoErr.BeginTxError)
			},
			expectedID:    0,
			expectedError: repoErr.BeginTxError,
		},
		{
			name:   "Faild commit transaction",
			params: &repoErraultParams,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectQuery("INSERT INTO photos").
					WithArgs("user-uuid", "test.png", sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows(idColumn).
						AddRow(1))

				mock.ExpectExec("INSERT INTO photo_versions").
					WithArgs(1, "home/user-uuid/test.png", 12345, 100, 100, sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit().WillReturnError(repoErr.CommitTxError)
			},
			expectedID:    0,
			expectedError: repoErr.CommitTxError,
		},
		{
			name:   "Failed insert photo",
			params: &repoErraultParams,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectQuery("INSERT INTO photos").
					WithArgs("user-uuid", "test.png", sqlmock.AnyArg()).
					WillReturnError(repoErr.InsertError)

				mock.ExpectRollback()
			},
			expectedID:    0,
			expectedError: repoErr.InsertError,
		},
		{
			name:   "Failed insert version",
			params: &repoErraultParams,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectQuery("INSERT INTO photos").
					WithArgs("user-uuid", "test.png", sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows(idColumn).
						AddRow(1))

				mock.ExpectExec("INSERT INTO photo_versions").
					WithArgs(1, "home/user-uuid/test.png", 12345, 100, 100, sqlmock.AnyArg()).
					WillReturnError(repoErr.InsertError)

				mock.ExpectRollback()
			},
			expectedID:    0,
			expectedError: repoErr.InsertError,
		},
		{
			name:   "Correct ID returned",
			params: &repoErraultParams,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery("INSERT INTO photos").
					WithArgs("user-uuid", "test.png", sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows(idColumn).AddRow(123))
				mock.ExpectExec("INSERT INTO photo_versions").
					WithArgs(123, "home/user-uuid/test.png", 12345, 100, 100, sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedID:    123,
			expectedError: nil,
		},
		{
			name:          "Nil params",
			params:        nil,
			mockSetup:     func(mock sqlmock.Sqlmock) {},
			expectedID:    0,
			expectedError: repoErr.NilParamsError,
		},
		{
			name:   "Invalid params",
			params: &model.CreateOriginalPhotoParams{},
			mockSetup: func(mock sqlmock.Sqlmock) {
			},
			expectedID:    0,
			expectedError: repoErr.InvalidParamsError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			repo := NewRepository(sqlxDB)

			tt.mockSetup(mock)

			photoID, err := repo.CreateOriginalPhoto(context.Background(), tt.params)
			fmt.Println(err)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, photoID)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_GetPhotoByID(t *testing.T) {
	uploadedAt := sql.NullTime{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}

	query := "SELECT id, user_uuid, filename, uploaded_at FROM photos WHERE id = \\$1"

	tests := []struct {
		name           string
		photoID        int
		mockSetup      func(mock sqlmock.Sqlmock)
		expectedResult *model.Photo
		expectedError  error
	}{
		{
			name:    "Valid",
			photoID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(photoColumns).
						AddRow(1, "user-uuid", "test.png", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)))
			},
			expectedResult: &model.Photo{
				ID:         1,
				UserUUID:   "user-uuid",
				Filename:   "test.png",
				UploadedAt: &uploadedAt,
			},
			expectedError: nil,
		},
		{
			name:    "Not found",
			photoID: 42,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs(42).
					WillReturnError(sql.ErrNoRows)
			},
			expectedResult: nil,
			expectedError:  repoErr.NotFoundError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			repo := NewRepository(sqlxDB)

			tt.mockSetup(mock)

			result, err := repo.PhotoByID(context.Background(), tt.photoID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_GetPhotoVersions(t *testing.T) {
	uploadedAt := sql.NullTime{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}
	query := "SELECT id, photo_id, version_type, uuid_filename, size, height, width, saved_at FROM photo_versions WHERE photo_id = \\$1 ORDER BY size"

	tests := []struct {
		name           string
		photoID        int
		mockSetup      func(mock sqlmock.Sqlmock)
		expectedResult []model.PhotoVersion
		expectedError  error
	}{
		{
			name:    "Valid",
			photoID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(photoVersionColumns).
						AddRow(1, 1, "original", "uuid_filename", 12345, 100, 100, uploadedAt))
			},
			expectedResult: []model.PhotoVersion{
				{ID: 1, PhotoID: 1,
					VersionType:  sql.NullString{String: "original", Valid: true},
					UUIDFilename: "uuid_filename",
					Size:         12345,
					Height:       100, Width: 100, SavedAt: &sql.NullTime{Time: uploadedAt.Time, Valid: true},
				},
			},
			expectedError: nil,
		},
		{
			name:    "Select error",
			photoID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs(1).
					WillReturnError(errors.New("select error"))
			},
			expectedResult: nil,
			expectedError:  errors.New("select error"),
		},
		{
			name:    "Empty result",
			photoID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(photoVersionColumns))
			},
			expectedResult: nil,
			expectedError:  nil,
		},
		{
			name:    "Photo not found",
			photoID: 42,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs(42).
					WillReturnError(sql.ErrNoRows)
			},
			expectedResult: nil,
			expectedError:  repoErr.NotFoundError,
		},
		{
			name:    "No versions of photo",
			photoID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs(10).
					WillReturnRows(sqlmock.NewRows(photoVersionColumns))
			},
			expectedResult: nil,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			repo := NewRepository(sqlxDB)

			tt.mockSetup(mock)

			versions, err := repo.GetPhotoVersions(context.Background(), tt.photoID)
			log.Warnf("error : %v", err)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, versions)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_GetPhotoVersionByToken(t *testing.T) {
	uploadedAt := sql.NullTime{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}
	query := `
	SELECT pv.id, pv.photo_id, pv.version_type, pv.uuid_filename, pv.size, pv.height, pv.width, pv.saved_at
	FROM published_photo_info ppi
	JOIN photo_versions pv ON ppi.photo_id = pv.photo_id
	WHERE ppi.public_token = ? AND version_type = ?`

	tests := []struct {
		name           string
		token          string
		version        domainModel.PhotoVersionType
		mockSetup      func(mock sqlmock.Sqlmock)
		expectedResult *model.PhotoVersion
		expectedError  error
	}{
		{
			name:    "Valid",
			token:   "token",
			version: domainModel.Original,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs("token", "original").
					WillReturnRows(sqlmock.NewRows(photoVersionColumns).AddRow(
						1, 1, "original", "uuid_filename1", int64(12345), 100, 100, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					))
			},
			expectedResult: &model.PhotoVersion{
				ID:           1,
				PhotoID:      1,
				VersionType:  sql.NullString{String: "original", Valid: true},
				UUIDFilename: "uuid_filename1",
				Size:         12345,
				Height:       100,
				Width:        100,
				SavedAt:      &sql.NullTime{Time: uploadedAt.Time, Valid: true},
			},
			expectedError: nil,
		},
		{
			name:    "Select error",
			token:   "token",
			version: domainModel.Original,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs("token", "original").
					WillReturnError(errors.New("select error"))
			},
			expectedResult: nil,
			expectedError:  errors.New("select error"),
		},
		{
			name:    "Empty result",
			token:   "token",
			version: domainModel.Original,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs("token", "original").
					WillReturnRows(sqlmock.NewRows(photoVersionColumns))
			},
			expectedResult: nil,
			expectedError:  repoErr.NotFoundError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			repo := NewRepository(sqlxDB)

			tt.mockSetup(mock)

			version, err := repo.GetPhotoVersionByToken(context.Background(), tt.token, &model.FilterParams{
				VersionType: tt.version,
			})
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, version)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_CreatePhotoPublishedInfo(t *testing.T) {
	type mockBehavior func(mock sqlmock.Sqlmock, photoID int)

	query := `
	INSERT INTO published_photo_info (photo_id)
	VALUES ($1)
	RETURNING public_token`

	returnedRow := sqlmock.NewRows([]string{"public_token"})

	tests := []struct {
		name          string
		photoID       int
		mockBehavior  mockBehavior
		expectedToken string
		expectedError error
	}{
		{
			name:    "Valid",
			photoID: 1,
			mockBehavior: func(mock sqlmock.Sqlmock, photoID int) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(photoID).
					WillReturnRows(returnedRow.AddRow("some-token"))
			},
			expectedToken: "some-token",
			expectedError: nil,
		},
		{
			name:    "Insert error",
			photoID: 1,
			mockBehavior: func(mock sqlmock.Sqlmock, photoID int) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(photoID).
					WillReturnError(errors.New("insert error"))
			},
			expectedToken: "",
			expectedError: errors.New("insert error"),
		},
		{
			name:    "Already exists",
			photoID: 1,
			mockBehavior: func(mock sqlmock.Sqlmock, photoID int) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(photoID).
					WillReturnError(
						&pq.Error{Code: pkgRepo.UniqueViolationErrorCode},
					)
			},
			expectedToken: "",
			expectedError: repoErr.ConflictError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			repo := NewRepository(sqlxDB)

			tt.mockBehavior(mock, tt.photoID)

			token, err := repo.CreatePhotoPublishedInfo(context.Background(), tt.photoID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_GetPublicPhotosByTokenPrefix(t *testing.T) {
	uploadedAt := sql.NullTime{
		Time:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}

	tests := []struct {
		name           string
		tokenPrefix    string
		filterParams   *model.FilterParams
		mockSetup      func(mock sqlmock.Sqlmock)
		expectedResult []model.PhotoWithPhotoVersion
		expectedError  error
	}{
		{
			name:         "Valid",
			tokenPrefix:  "token",
			filterParams: nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				queryRebinded := `
	SELECT
    	p.id AS photo_id,
    	p.user_uuid,
    	p.filename,
    	p.uploaded_at,
    	pv.id AS version_id,
    	pv.version_type,
    	pv.size,
    	pv.width,
    	pv.uuid_filename,
    	pv.height,
    	pv.saved_at
	FROM photos p
	INNER JOIN published_photo_info pi
    	ON p.id = pi.photo_id
	INNER JOIN photo_versions pv
        ON p.id = pv.photo_id
	WHERE pi.public_token LIKE ?`

				mock.ExpectQuery(regexp.QuoteMeta(queryRebinded)).
					WithArgs("token%").
					WillReturnRows(sqlmock.NewRows([]string{
						"photo_id", "user_uuid", "filename", "uploaded_at",
						"version_id", "version_type", "size", "width", "uuid_filename", "height", "saved_at",
					}).AddRow(
						1, "user-uuid", "test.png", uploadedAt.Time,
						2, "original", 1234, 100, "uuid_file.png", 200, uploadedAt.Time,
					))
			},
			expectedResult: []model.PhotoWithPhotoVersion{
				{
					PhotoID:      1,
					UserUUID:     "user-uuid",
					Filename:     "test.png",
					UploadedAt:   &uploadedAt,
					VersionID:    2,
					VersionType:  sql.NullString{String: "original", Valid: true},
					Size:         1234,
					Width:        100,
					UUIDFilename: "uuid_file.png",
					Height:       200,
					SavedAt:      &uploadedAt,
				},
			},
			expectedError: nil,
		},
		{
			name:        "With FilterParams",
			tokenPrefix: "abc",
			filterParams: &model.FilterParams{
				VersionType: "original",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				queryRebinded := `
	SELECT
    	p.id AS photo_id,
    	p.user_uuid,
    	p.filename,
    	p.uploaded_at,
    	pv.id AS version_id,
    	pv.version_type,
    	pv.size,
    	pv.width,
    	pv.uuid_filename,
    	pv.height,
    	pv.saved_at
	FROM photos p
	INNER JOIN published_photo_info pi
    	ON p.id = pi.photo_id
	INNER JOIN photo_versions pv
        ON p.id = pv.photo_id
	WHERE pi.public_token LIKE ? AND version_type = ?`

				mock.ExpectQuery(regexp.QuoteMeta(queryRebinded)).
					WithArgs("abc%", "original").
					WillReturnRows(sqlmock.NewRows([]string{
						"photo_id", "user_uuid", "filename", "uploaded_at",
						"version_id", "version_type", "size", "width", "uuid_filename", "height", "saved_at",
					}).AddRow(
						2, "another-user", "img.jpg", uploadedAt.Time,
						5, "original", 999, 300, "another_file.jpg", 400, uploadedAt.Time,
					))
			},
			expectedResult: []model.PhotoWithPhotoVersion{
				{
					PhotoID:      2,
					UserUUID:     "another-user",
					Filename:     "img.jpg",
					UploadedAt:   &uploadedAt,
					VersionID:    5,
					VersionType:  sql.NullString{String: "original", Valid: true},
					Size:         999,
					Width:        300,
					UUIDFilename: "another_file.jpg",
					Height:       400,
					SavedAt:      &uploadedAt,
				},
			},
			expectedError: nil,
		},
		{
			name:         "Empty result",
			tokenPrefix:  "notfound",
			filterParams: nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				queryRebinded := `
	SELECT
    	p.id AS photo_id,
    	p.user_uuid,
    	p.filename,
    	p.uploaded_at,
    	pv.id AS version_id,
    	pv.version_type,
    	pv.size,
    	pv.width,
    	pv.uuid_filename,
    	pv.height,
    	pv.saved_at
	FROM photos p
	INNER JOIN published_photo_info pi
    	ON p.id = pi.photo_id
	INNER JOIN photo_versions pv
        ON p.id = pv.photo_id
	WHERE pi.public_token LIKE ?`

				mock.ExpectQuery(regexp.QuoteMeta(queryRebinded)).
					WithArgs("notfound%").
					WillReturnRows(sqlmock.NewRows([]string{
						"photo_id", "user_uuid", "filename", "uploaded_at",
						"version_id", "version_type", "size", "width", "uuid_filename", "height", "saved_at",
					}))
			},
			expectedResult: nil,
			expectedError:  nil,
		},
		{
			name:         "Query error",
			tokenPrefix:  "fail",
			filterParams: nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				queryRebinded := `
	SELECT
    	p.id AS photo_id,
    	p.user_uuid,
    	p.filename,
    	p.uploaded_at,
    	pv.id AS version_id,
    	pv.version_type,
    	pv.size,
    	pv.width,
    	pv.uuid_filename,
    	pv.height,
    	pv.saved_at
	FROM photos p
	INNER JOIN published_photo_info pi
    	ON p.id = pi.photo_id
	INNER JOIN photo_versions pv
        ON p.id = pv.photo_id
	WHERE pi.public_token LIKE ?`

				mock.ExpectQuery(regexp.QuoteMeta(queryRebinded)).
					WithArgs("fail%").
					WillReturnError(fmt.Errorf("db error"))
			},
			expectedResult: nil,
			expectedError:  fmt.Errorf("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			repo := NewRepository(sqlxDB)

			tt.mockSetup(mock)

			result, err := repo.GetPublicPhotosByTokenPrefix(context.Background(), tt.tokenPrefix, tt.filterParams)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepository_DeletePhotoPublishedInfo(t *testing.T) {
	type mockBehavior func(mock sqlmock.Sqlmock, photoID int)

	query := `
	DELETE FROM published_photo_info
	WHERE photo_id = $1`

	tests := []struct {
		name          string
		photoID       int
		mockBehavior  mockBehavior
		expectedError error
	}{
		{
			name:    "Valid",
			photoID: 1,
			mockBehavior: func(mock sqlmock.Sqlmock, photoID int) {
				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(photoID).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			repo := NewRepository(sqlxDB)

			tt.mockBehavior(mock, tt.photoID)

			err = repo.DeletePhotoPublishedInfo(context.Background(), tt.photoID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}
