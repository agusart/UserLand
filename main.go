package main

import (
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"net/http"
)


func main(){
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file")
	}

	srv := InitServer()
	_ = http.ListenAndServe(":8080", srv)
}