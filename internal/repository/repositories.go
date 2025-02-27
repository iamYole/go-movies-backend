package repository

import (
	"context"
	"database/sql"

	"github.com/iamYole/go-movies/internal/models"
)

type Repository struct {
	Movies interface {
		GetMovies(context.Context) ([]*models.Movie, error)
		GetMovieByID(context.Context, int64) (*models.Movie, error)
		EditMovie(context.Context, int64) (*models.Movie,[]*models.Genre, error)
	}
	Users interface {
		GetUserByEmail(context.Context, string) (*models.User, error)
		GetUserByID(context.Context, int64)(*models.User, error)
		CreateUser(context.Context, models.User) error
	}
}

func NewDbConn(db *sql.DB) Repository {
	return Repository{
		Movies: &models.MovieRepo{DB: db},
		Users:  &models.UserRepo{DB: db},
	}
}
