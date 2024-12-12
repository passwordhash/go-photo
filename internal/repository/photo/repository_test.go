package photo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	def "go-photo/internal/repository"
	"go-photo/internal/repository/photo/model"
	"testing"
	"time"
)

var (
	photoColumns        = []string{"id", "user_uuid", "filename", "uploaded_at"}
	photoVersionColumns = []string{"id", "photo_id", "version_type", "filepath", "size"}
	idColumn            = []string{"id"}
)

var (
	rowsWithPhotoColumns        = sqlmock.NewRows(photoColumns)
	rowsWithPhotoVersionColumns = sqlmock.NewRows(photoVersionColumns)
	rowsWithIDColumn            = sqlmock.NewRows(idColumn)
)

func TestRepository_CreateOriginalPhoto(t *testing.T) {
	defaultParams := model.CreateOriginalPhotoParams{
		UserUUID: "user-uuid",
		Filename: "test.png",
		Filepath: "home/user-uuid/test.png",
		Size:     12345,
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
			params: &defaultParams,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectQuery("INSERT INTO photos").
					WithArgs("user-uuid", "test.png").
					WillReturnRows(rowsWithIDColumn.
						AddRow(1))

				mock.ExpectExec("INSERT INTO photo_versions").
					WithArgs(1, "home/user-uuid/test.png", 12345).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
			expectedID:    1,
			expectedError: nil,
		},
		{
			name:   "Failed begin transaction",
			params: &defaultParams,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(def.BeginTxError)
			},
			expectedID:    0,
			expectedError: def.BeginTxError,
		},
		{
			name:   "Faild commit transaction",
			params: &defaultParams,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectQuery("INSERT INTO photos").
					WithArgs("user-uuid", "test.png").
					WillReturnRows(rowsWithIDColumn.
						AddRow(1))

				mock.ExpectExec("INSERT INTO photo_versions").
					WithArgs(1, "home/user-uuid/test.png", 12345).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit().WillReturnError(def.CommitTxError)
			},
			expectedID:    0,
			expectedError: def.CommitTxError,
		},
		{
			name:   "Failed insert photo",
			params: &defaultParams,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectQuery("INSERT INTO photos").
					WithArgs("user-uuid", "test.png").
					WillReturnError(def.InsertPhotoError)

				mock.ExpectRollback()
			},
			expectedID:    0,
			expectedError: def.InsertPhotoError,
		},
		{
			name:   "Failed insert version",
			params: &defaultParams,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectQuery("INSERT INTO photos").
					WithArgs("user-uuid", "test.png").
					WillReturnRows(rowsWithIDColumn.
						AddRow(1))

				mock.ExpectExec("INSERT INTO photo_versions").
					WithArgs(1, "home/user-uuid/test.png", 12345).
					WillReturnError(def.InsertVersionError)

				mock.ExpectRollback()
			},
			expectedID:    0,
			expectedError: def.InsertVersionError,
		},
		{
			name:   "Correct ID returned",
			params: &defaultParams,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery("INSERT INTO photos").
					WithArgs("user-uuid", "test.png").
					WillReturnRows(rowsWithIDColumn.AddRow(123))
				mock.ExpectExec("INSERT INTO photo_versions").
					WithArgs(123, "home/user-uuid/test.png", 12345).
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
			expectedError: def.NilParamsError,
		},
		{
			name:   "Invalid params",
			params: &model.CreateOriginalPhotoParams{},
			mockSetup: func(mock sqlmock.Sqlmock) {
			},
			expectedID:    0,
			expectedError: def.InvalidParamsError,
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

func TestRepository_GetPhotoVersions(t *testing.T) {
	uploadedAt := sql.NullTime{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}

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
				mock.ExpectQuery("SELECT id, user_uuid, filename, uploaded_at FROM photos WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(photoColumns).
						AddRow(1, "user-uuid", "photo.jpg", uploadedAt))

				mock.ExpectQuery("SELECT id, photo_id, version_type, filepath, size FROM photo_versions WHERE photo_id = \\$1 ORDER BY size").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(photoVersionColumns).
						AddRow(1, 1, "original", "filepath1", 12345).
						AddRow(2, 1, "thumbnail", "filepath2", 54321))
			},
			expectedResult: []model.PhotoVersion{
				{ID: 1, PhotoID: 1, VersionType: sql.NullString{String: "original", Valid: true}, Filepath: "filepath1", Size: 12345},
				{ID: 2, PhotoID: 1, VersionType: sql.NullString{String: "thumbnail", Valid: true}, Filepath: "filepath2", Size: 54321},
			},
			expectedError: nil,
		},
		{
			name:    "Select error",
			photoID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, user_uuid, filename, uploaded_at FROM photos WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(photoColumns).
						AddRow(1, "user-uuid", "photo.jpg", uploadedAt))

				mock.ExpectQuery("SELECT id, photo_id, version_type, filepath, size FROM photo_versions WHERE photo_id = \\$1 ORDER BY size").
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
				mock.ExpectQuery("SELECT id, user_uuid, filename, uploaded_at FROM photos WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(photoColumns).
						AddRow(1, "user-uuid", "photo.jpg", uploadedAt))

				mock.ExpectQuery("SELECT id, photo_id, version_type, filepath, size FROM photo_versions WHERE photo_id = \\$1 ORDER BY size").
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
				mock.ExpectQuery("SELECT id, user_uuid, filename, uploaded_at FROM photos WHERE id = \\$1").
					WithArgs(42).
					WillReturnError(sql.ErrNoRows)
			},
			expectedResult: nil,
			expectedError:  def.PhotoNotFound,
		},
		{
			name:    "No versions of photo",
			photoID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, user_uuid, filename, uploaded_at FROM photos WHERE id = \\$1").
					WithArgs(10).
					WillReturnRows(sqlmock.NewRows(photoColumns).
						AddRow(10, "user-uuid", "photo.jpg", uploadedAt))

				mock.ExpectQuery("SELECT id, photo_id, version_type, filepath, size FROM photo_versions WHERE photo_id = \\$1 ORDER BY size").
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
