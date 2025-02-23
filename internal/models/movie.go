package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/iamYole/go-movies/internal/db"
)

type Movie struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	ReleaseDate time.Time `json:"release_date"`
	Runtime     int       `json:"runtime"`
	MPAARating  string    `json:"mpaa_rating"`
	Description string    `json:"descript"`
	Image       string    `json:"string"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

type MovieRepo struct {
	DB *sql.DB
}

func (m *MovieRepo) GetMovies(ctx context.Context) ([]*Movie, error) {
	var movies []*Movie
	qry := `select 
				m.id, m.title, m.release_date, m.runtime, m.mpaa_rating,
				m.description ,coalesce(m.image,'') ,m.created_at ,m.updated_at 
			from 
				movies m
			order by m.title;`

	ctx, cancel := context.WithTimeout(ctx, db.QueryTimeoutDuration)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, qry)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m Movie
		err := rows.Scan(
			&m.ID,
			&m.Title,
			&m.ReleaseDate,
			&m.Runtime,
			&m.MPAARating,
			&m.Description,
			&m.Image,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		movies = append(movies, &m)
	}

	return movies, nil
}

func (m *MovieRepo) GetMovieByID(ctx context.Context, movieID int64) (Movie, error) {
	var movie Movie

	return movie, nil
}
