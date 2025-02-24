package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/iamYole/go-movies/internal/models"
)

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	var payLoad = struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Version string `json:"version"`
	}{
		Status:  "active",
		Message: "Server is up and running",
		Version: "1.0.0",
	}

	err := app.WriteJSON(w, http.StatusOK, payLoad)
	if err != nil {
		app.WriteJSONError(w, err, http.StatusInternalServerError)
	}
}

func (app *application) AllMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := app.repo.Movies.GetMovies(r.Context())
	if err != nil {
		app.WriteJSONError(w, err)
		return
	}

	err = app.WriteJSON(w, http.StatusOK, movies)
	if err != nil {
		app.WriteJSONError(w, err, http.StatusInternalServerError)
	}

}

func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {
	var loginPayload struct {
		Email    string `json:"email" validate:"required,email,max=255"`
		Password string `json:"password" validate:"required,min=3,max=50"`
	}
	
	//read JSON payload
	err := app.ReadJSON(w, r, &loginPayload)
	if err != nil {
		app.WriteJSONError(w, err)
		return
	}

	if err := Validate.Struct(loginPayload);err!=nil{
		app.WriteJSONError(w,errors.New("please fill in all required fields"))
		return
	}

	//validate user against the database
	user, err := app.repo.Users.GetUserByEmail(r.Context(), loginPayload.Email)
	if err != nil {
		app.WriteJSONError(w, errors.New("invalid credentials"))
		return
	}
	valid, err := user.Password.ValidatePassword(loginPayload.Password)
	if err != nil || !valid {
		app.WriteJSONError(w, errors.New("invalid credentials"))
		log.Println(err)
		return
	}

	//generate tokens
	tokens := app.generateAndSendToken(w,user)

	app.WriteJSON(w, http.StatusAccepted, tokens.Token)
}

type CreateUserPayload struct {
	FirstName string `json:"first_name" validate:"required,max=50"`
	LastName  string `json:"last_name" validate:"required,max=50"`
	Email     string `json:"email" validate:"required,email,max=50"`
	Password  string `json:"password" validate:"required,min=3,max=50"`
}

func (app *application) Register(w http.ResponseWriter, r *http.Request) {
	var payload CreateUserPayload

	if err := app.ReadJSON(w, r, &payload); err != nil {
		app.WriteJSONError(w, err)
		return
	}

	if err:=Validate.Struct(payload);err!=nil{
		app.WriteJSONError(w,errors.New("please fill in all required fields"))
	}

	user := &models.User{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
	}

	if err := user.Password.Set(payload.Password); err != nil {
		app.WriteJSONError(w, err, http.StatusInternalServerError)
		return
	}

	if err := app.repo.Users.CreateUser(r.Context(), *user); err != nil {
		app.WriteJSONError(w, err)
		return
	}

	tokens := app.generateAndSendToken(w,user)
	

	if err := app.WriteJSON(w, http.StatusCreated, tokens.Token); err != nil {
		app.WriteJSONError(w, err, http.StatusInternalServerError)
		return
	}
}

func (app *application) generateAndSendToken(w http.ResponseWriter, user *models.User)*TokenPairs{
		//create jwtuser
		u := jwtUser{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		}
	
		//generate tokens
		tokens, err := app.auth.GeneratToken(&u)
		if err != nil {
			app.WriteJSONError(w, err)
			return nil
		}
	
		refreshCookie := app.auth.GetRefreshCookie(tokens.RefreshToken)
		http.SetCookie(w, refreshCookie)
		return &tokens
}
