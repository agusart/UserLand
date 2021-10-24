package main

import (
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"net/http"
	"os"
	"strconv"
	"userland/store/postgres"
)


func main(){	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	err := godotenv.Load()
	if err != nil {
		log.Err(err)
	}


	port, err := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	if err != nil {
		log.Err(err)
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
		log.Err( err)
	}


	srv := InitServer(db)
	_ = http.ListenAndServe(":8080", srv)
}