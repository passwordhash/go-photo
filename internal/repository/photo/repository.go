package photo

import (
	def "go-photo/internal/repository"

	"github.com/jmoiron/sqlx"
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
