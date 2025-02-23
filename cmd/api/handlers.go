package main

import (
	"net/http"
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
