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

func (r *repository) GetPhotoVersions(_ context.Context, photoID int) ([]repoModel.PhotoVersion, error) {
	var versions []repoModel.PhotoVersion

	query := `
		SELECT id, photo_id, version_type, filepath, width, height, size 
		FROM photo_versions 
		WHERE photo_id = $1`

	err := r.db.Select(&versions, query, photoID)
	if err != nil {
		return nil, err
	}

	fmt.Println("Photo versions: ", versions)

	return versions, err
}
