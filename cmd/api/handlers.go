package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/iamYole/go-movies/internal/models"
	//"github.com/iamYole/go-movies/internal/models"
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
		app.WriteJSONError(w, err,http.StatusInternalServerError)
		return
	}

	err = app.WriteJSON(w, http.StatusOK, movies)
	if err != nil {
		app.WriteJSONError(w, err, http.StatusInternalServerError)
	}

}

func (app *application) GetAllGenresHandle(w http.ResponseWriter, r *http.Request){
	genres, err := app.repo.Movies.GetAllGenres(r.Context())
	if err!=nil{
		app.WriteJSONError(w,err,http.StatusInternalServerError)
		return
	}

	if err := app.WriteJSON(w,http.StatusOK,genres);err!=nil{
		app.WriteJSONError(w,err,http.StatusInternalServerError)
		return
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

// type MoviePayload struct{
// 	//ID          int       `json:"id" validate:"required"`
// 	Title       string    `json:"title" validate:"required"`
// 	ReleaseDate time.Time `json:"release_date" validate:"required"`
// 	Runtime     int       `json:"runtime" validate:"required"`
// 	MPAARating  string    `json:"mpaa_rating" validate:"required"`
// 	Description string    `json:"description" validate:"required"`
// 	Image       string    `json:"image" validate:"required"`
// 	//CreatedAt   time.Time `json:"-" validate:"required"`
// 	//UpdatedAt   time.Time `json:"-"`
// 	//Genres      []int  `json:"genres,validate:"required"`
// 	GenresArray []int     `json:"genres_array,omitempty"`
// }
func (app *application) InsertMovieHandler(w http.ResponseWriter, r *http.Request){
	var movie models.Movie

	if err:= app.ReadJSON(w,r,&movie);err!=nil{
		app.WriteJSONError(w,err)
		return
	}

	// get image
	movie = app.getPoster(movie)

	newID,err := app.repo.Movies.InsertMovie(r.Context(),movie)
	if err!=nil{
		app.WriteJSONError(w,err,http.StatusInternalServerError)
		return
	}

	//handle genre
	err = app.repo.Movies.UpdateMovieGenres(r.Context(),int(newID),movie.GenresArray)
	if err!=nil{
		app.WriteJSONError(w,err,http.StatusInternalServerError)
		return
	}

	res := JSONResponse{
		Error: false,
		Message: "Movie Added",
	}
	if err:=app.WriteJSON(w,http.StatusCreated,res);err!=nil{
		app.WriteJSONError(w,err,http.StatusInternalServerError)
		return
	}
}

func (app *application) MovieCatalog(w http.ResponseWriter, r *http.Request){
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

func (app *application) GetMovieHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	movieID, err := strconv.Atoi(id)
	if err != nil {
		app.WriteJSONError(w, err)
		return
	}

	movie, err := app.repo.Movies.GetMovieByID(r.Context(),int64(movieID))
	if err != nil {
		app.WriteJSONError(w, err)
		return
	}

	if err := app.WriteJSON(w, http.StatusOK, movie); err!=nil{
		app.WriteJSONError(w,err,http.StatusInternalServerError)
	}
}

func (app *application) EditMovieHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	movieID, err := strconv.Atoi(id)
	if err != nil {
		app.WriteJSONError(w, err)
		return
	}

	movie, genres, err := app.repo.Movies.EditMovie(r.Context(), int64(movieID))
	if err != nil {
		app.WriteJSONError(w, err)
		return
	}

	var payload = struct {
		Movie  *models.Movie   `json:"movie"`
		Genres []*models.Genre `json:"genres"`
	}{
		movie,
		genres,
	}

	if err := app.WriteJSON(w, http.StatusOK, payload);err !=nil{
		app.WriteJSONError(w,err,http.StatusInternalServerError)
	}
}


func (app *application)refreshToken(w http.ResponseWriter, r *http.Request){
	for _, cookie := range r.Cookies(){
		if cookie.Name == app.auth.CookieName{
			claims := &Claims{}
			refreshToken := cookie.Value

			//parse the token to claims
			_, err := jwt.ParseWithClaims(refreshToken,claims,func(t *jwt.Token) (interface{}, error) {
				return []byte(app.cfg.authCfg.JWTSecret), nil
			})
			if err !=nil{
				app.WriteJSONError(w,errors.New("unauthorised"),http.StatusUnauthorized)
				return
			}

			//get userid from token claims
			userID,err := strconv.Atoi(claims.Subject)
			if err !=nil{
				app.WriteJSONError(w,errors.New("unknown user"),http.StatusUnauthorized)
				return
			}

			user, err := app.repo.Users.GetUserByID(r.Context(),int64(userID))
			if err!=nil{
				app.WriteJSONError(w,err,http.StatusInternalServerError)
				//log.Println(err)
				return
			}

			tokens := app.generateAndSendToken(w,user)

			if err := app.WriteJSON(w, http.StatusOK, tokens); err != nil {
				app.WriteJSONError(w, err, http.StatusInternalServerError)
				return
			}

		}
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

func(app *application) logout(w http.ResponseWriter, r *http.Request){
	http.SetCookie(w,app.auth.GetExpiredRefereshToken())
	w.WriteHeader(http.StatusAccepted)
}

func (app *application)getPoster(movie models.Movie)models.Movie{
	type TheMovieDB struct{
		Page int `json:"page"`
		Results []struct{
			PosterPath string `json:"poster_path"`
		}`json:"results"`
		TotalPages int `json:"total_pages"`
	}

	client := &http.Client{}
	theURL := app.imdb.search_url

	req, err := http.NewRequest("GET",theURL+"&query="+url.QueryEscape(movie.Title),nil)
	if err!=nil{
		//log.Println(theURL+"&query="+url.QueryEscape(movie.Title))
		log.Println(err)
		return movie
	}

	req.Header.Add("Accept","application/json")
	req.Header.Add("Content-Type","application/json")

	res, err := client.Do(req)
	if err!=nil{
		log.Println(err)
		return movie
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err!=nil{
		log.Println(err)
		return movie
	}

	var responseObj TheMovieDB
	json.Unmarshal(bodyBytes,&responseObj)

	if len(responseObj.Results) > 0{
		movie.Image = responseObj.Results[0].PosterPath
	}

	return movie
}