package main

import (
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
	"userland/store/postgres"
)


func main(){
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	port, err := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	if err != nil {
		log.Fatalf("cant convert db port env to integer: %v", err)
	}
	postgresCfg := postgres.PGConfig{
		Host: os.Getenv("POSTGRES_ADDR"),
		Port: port,
		Username: os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Database: os.Getenv("POSTGRES_DB"),
	}

	db, err :=  postgres.NewPG(postgresCfg)
	if err != nil {
		log.Fatalf("cant open db: %v", err)
	}


	srv := InitServer(db)
	_ = http.ListenAndServe(":8080", srv)
}