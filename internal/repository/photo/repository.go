package photo

import (
	"github.com/jmoiron/sqlx"
	def "go-photo/internal/repository"
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
