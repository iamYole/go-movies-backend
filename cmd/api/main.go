package main

import (
	"fmt"
	"log"
	"net/http"
)

const port = 8080

type application struct{
	Domain string
	port int	
}

func main(){
	//set app config
	//read from command line
	//connect to the database
	//start a webserver
	var app application
	app.Domain = "example.com"
	app.port = 8080


	log.Println("Startng server on port ", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d",app.port),app.routes())
	if err!=nil{
		log.Fatal(err)
	}
}