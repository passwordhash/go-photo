package photo

import (
	"context"
	"database/sql"
	"errors"
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
)

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
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_uuid", "filename", "uploaded_at"}).
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
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_uuid", "filename", "uploaded_at"}).
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
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_uuid", "filename", "uploaded_at"}).
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
			name:    "Photo ID mismatch",
			photoID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, user_uuid, filename, uploaded_at FROM photos WHERE id = \\$1").
					WithArgs(10).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_uuid", "filename", "uploaded_at"}).
						AddRow(10, "user-uuid", "photo.jpg", uploadedAt))

				mock.ExpectQuery("SELECT id, photo_id, version_type, filepath, size FROM photo_versions WHERE photo_id = \\$1 ORDER BY size").
					WithArgs(10).
					WillReturnRows(sqlmock.NewRows(photoVersionColumns).
						AddRow(1, 5, "original", "filepath1", 12345))
			},
			expectedResult: []model.PhotoVersion{
				{ID: 1, PhotoID: 5, VersionType: sql.NullString{String: "original", Valid: true}, Filepath: "filepath1", Size: 12345},
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
