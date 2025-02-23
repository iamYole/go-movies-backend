package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/iamYole/go-movies/internal/db"
	"github.com/iamYole/go-movies/internal/env"
	"github.com/iamYole/go-movies/internal/repository"
)

const port = 8080

type application struct {
	Domain string
	cfg    config
	repo   repository.Repository
}
type config struct {
	port int
	dsn  dbconnection
}
type dbconnection struct {
	dsn string
}

func main() {
	cfg := &config{
		port: env.GetInt("PORT", 8080),
		dsn: dbconnection{
			dsn: env.GetString("DSN", "dsn"),
		},
	}

	//connect to database
	db, err := db.New(cfg.dsn.dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("database connection established")

	repo := repository.NewDbConn(db)

	app := &application{
		Domain: env.GetString("DOMAIN", "example.com"),
		cfg:    *cfg,
		repo:   repo,
	}

	log.Println("Startng server on port ", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", app.cfg.port), app.routes())
	if err != nil {
		log.Fatal(err)
	}
}
