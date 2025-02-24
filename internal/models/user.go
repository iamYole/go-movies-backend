package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/iamYole/go-movies/internal/db"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

func (p *password) ValidatePassword(plainTextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plainTextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			//invalid password
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

type UserRepo struct {
	DB *sql.DB
}

func (u *UserRepo) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	qry := `select 
				u.id ,u.first_name, u.last_name, u.email ,u."password" ,u.created_at ,u.updated_at  
			from users u
			where u.email = $1;`

	ctx, cancel := context.WithTimeout(ctx, db.QueryTimeoutDuration)
	defer cancel()

	row := u.DB.QueryRowContext(ctx, qry, email)
	err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserRepo) CreateUser(ctx context.Context, user User) error {
	qry := `insert into users (first_name, last_name, email,password,created_at, updated_at)
			values($1,$2,$3,$4,$5,$6) RETURNING id, created_at;`

	ctx, cancel := context.WithTimeout(ctx, db.QueryTimeoutDuration)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, qry, user.FirstName,
		user.LastName, user.Email, user.Password.hash, time.Now(), time.Now()).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}
