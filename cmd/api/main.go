package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/iamYole/go-movies/internal/db"
	"github.com/iamYole/go-movies/internal/env"
	"github.com/iamYole/go-movies/internal/repository"
)

const port = 8080

type application struct {
	Domain string
	cfg    config
	repo   repository.Repository
	auth   Authentication
}
type config struct {
	port    int
	dsn     dbconnection
	authCfg authConfig
}
type dbconnection struct {
	dsn string
}
type authConfig struct {
	JWTSecret     string
	JWTIssuer     string
	JWTAud        string
	CookieDomain  string
	TokenExpiry   int
	RefreshExpiry int
}

func main() {
	cfg := config{
		port: env.GetInt("PORT", 8080),
		dsn: dbconnection{
			dsn: env.GetString("DSN", "dsn"),
		},
		authCfg: authConfig{
			JWTSecret:     env.GetString("JWT_SECRET", "jwtsecret"),
			JWTIssuer:     env.GetString("JWT_ISSUER", "jwtiss"),
			JWTAud:        env.GetString("JWT_AUDIENCE", "jwtaud"),
			CookieDomain:  env.GetString("JWT_COOKIE_DOMAIN", "cookie_domain"),
			TokenExpiry:   env.GetInt("JWT_TOKEN_EXP", 15),                //15mins
			RefreshExpiry: env.GetInt("JWT_REFERESH_TOKEN_EXP", (24 * 7)), //7days
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
		cfg:    cfg,
		repo:   repo,
		auth: Authentication{
			Issuer:        cfg.authCfg.JWTIssuer,
			Audience:      cfg.authCfg.JWTAud,
			Secret:        cfg.authCfg.JWTSecret,
			TokenExpiry:   time.Minute * time.Duration(cfg.authCfg.TokenExpiry),
			RefreshExpiry: time.Hour * time.Duration(cfg.authCfg.RefreshExpiry),
			CookiePath:    "/",
			CookieName:    env.GetString("COOKIE_NAME", "__HOST-referesh_teken"),
			CookieDomain:  cfg.authCfg.CookieDomain,
		},
	}

	log.Println("Startng server on port ", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", app.cfg.port), app.routes())
	if err != nil {
		log.Fatal(err)
	}
}
