package repository

import (
	"context"
	"database/sql"

	"github.com/iamYole/go-movies/internal/models"
)

type Repository struct {
	Movies interface {
		GetMovies(context.Context) ([]*models.Movie, error)
		GetMovieByID(context.Context, int64) (models.Movie, error)
	}
}

func NewDbConn(db *sql.DB) Repository {
	return Repository{
		Movies: &models.MovieRepo{DB: db},
	}
}
