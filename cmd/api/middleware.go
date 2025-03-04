package main

import (
	"log"
	"net/http"
)

func (app *application) authRequired(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_,_,err := app.auth.GetTokenFromHeaderAndVerify(w,r)
		if err!=nil{
			w.WriteHeader(http.StatusUnauthorized)
			log.Println(err)
			return
		}
		next.ServeHTTP(w,r)
	})
}