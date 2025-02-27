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
	Description string    `json:"description"`
	Image       string    `json:"image"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
	Genres      []*Genre  `json:"genres,omitempty"`
	GenresArray []int     `json:"genres_array,omitempty"`
}

type Genre struct {
	ID        int       `json:"id"`
	Genre     string    `json:"genre"`
	Checked   bool      `json:"checked"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
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

func (m *MovieRepo) GetMovieByID(ctx context.Context, movieID int64) (*Movie, error) {
	var movie Movie
	qry := `select m.id, m.title, m.release_date,m.runtime,m.mpaa_rating ,m.description ,
				   coalesce(m.image,'') ,m.created_at ,m.updated_at 
			from movies m
			where m.id= $1;`

	ctx, cancel := context.WithTimeout(ctx, db.QueryTimeoutDuration)
	defer cancel()

	row := m.DB.QueryRowContext(ctx,qry,movieID)
	err:=row.Scan(
		&movie.ID,
		&movie.Title,
		&movie.ReleaseDate,
		&movie.Runtime,
		&movie.MPAARating,
		&movie.Description,
		&movie.Image,
		&movie.CreatedAt,
		&movie.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	qry = `select g.id, g.genre 
		   from movies_genres mg 
				left join genres g 
				on mg.id =g.id 
			where g.id =$1
			order by g.genre;`

	rows, err := m.DB.QueryContext(ctx,qry,movieID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	var genres []*Genre
	for rows.Next() {
		var g Genre
		err := rows.Scan(
			&g.ID,
			&g.Genre,
		)
		if err != nil {
			return nil, err
		}

		genres = append(genres, &g)
	}
	movie.Genres=genres

	return &movie, nil
}

func (m *MovieRepo) EditMovie(ctx context.Context, movieID int64) (*Movie, []*Genre, error) {
	var movie Movie
	qry := `select m.id, m.title, m.release_date,m.runtime,m.mpaa_rating ,m.description ,
				   coalesce(m.image,'') ,m.created_at ,m.updated_at 
			from movies m
			where m.id= $1;`

	ctx, cancel := context.WithTimeout(ctx, db.QueryTimeoutDuration)
	defer cancel()

	row := m.DB.QueryRowContext(ctx,qry,movieID)
	err:=row.Scan(
		&movie.ID,
		&movie.Title,
		&movie.ReleaseDate,
		&movie.Runtime,
		&movie.MPAARating,
		&movie.Description,
		&movie.Image,
		&movie.CreatedAt,
		&movie.UpdatedAt,
	)
	if err != nil {
		return nil,nil, err
	}

	qry = `select g.id, g.genre 
		   from movies_genres mg 
				left join genres g 
				on mg.id =g.id 
			where g.id =$1
			order by g.genre;`

	rows, err := m.DB.QueryContext(ctx,qry,movieID)
	if err != nil && err != sql.ErrNoRows {
		return nil,nil, err
	}
	defer rows.Close()

	var genres []*Genre
	var genresArray []int

	for rows.Next() {
		var g Genre
		err := rows.Scan(
			&g.ID,
			&g.Genre,
		)
		if err != nil {
			return nil,nil, err
		}

		genres = append(genres, &g)
		genresArray = append(genresArray, g.ID)
	}
	movie.Genres=genres
	movie.GenresArray = genresArray

	var allGenres []*Genre
	qry = "select id, genre, from genres order by genre"
	gRows, err := m.DB.QueryContext(ctx, qry)
	if err != nil {
		return nil, nil, err
	}
	defer gRows.Close()

	for gRows.Next() {
		var g Genre
		err := gRows.Scan(
			&g.ID,
			&g.Genre,
		)
		if err != nil {
			return nil, nil, err
		}

		allGenres = append(allGenres, &g)
	}

	return &movie, allGenres, nil
}

