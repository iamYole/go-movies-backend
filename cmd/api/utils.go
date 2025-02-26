package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New((validator.WithRequiredStructEnabled()))
}

type JSONResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (app *application) WriteJSON(w http.ResponseWriter, statusCode int, data any, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

func (app *application) ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1024 * 1024 //1mb
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decode := json.NewDecoder(r.Body)
	decode.DisallowUnknownFields()

	err := decode.Decode(data)
	if err != nil {
		return err
	}

	//ensures only one json file is read
	err = decode.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single json value")
	}

	return nil
}

func (app *application) WriteJSONError(w http.ResponseWriter, err error, staus ...int) error {
	statusCode := http.StatusBadRequest

	if len(staus) > 0 {
		statusCode = staus[0]
	}

	var payload JSONResponse
	payload.Error = true
	payload.Message = err.Error()

	return app.WriteJSON(w, statusCode, payload)
}
