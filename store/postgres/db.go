package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/rs/zerolog/log"
	"os"
)

var database *sql.DB

func NewDatabase() *sql.DB {
	url := getDatabaseUrl()
	return initDatabase(url)
}


func initDatabase(url string) *sql.DB {
	if database != nil {
		return database
	}

	db, err := sql.Open("pgx", url)
	if err != nil {
		log.Printf("could not connect to database: %v", err)
	}


	if err := db.Ping(); err != nil {
		log.Printf("unable to reach database: %v", err)
	}

	return db
}


func getDatabaseUrl() string {
	username := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	database := os.Getenv("POSTGRES_DB")
	network := os.Getenv("POSTGRES_ADDR")
	port := os.Getenv("POSTGRES_PORT")

	url := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s", username, password, network, port, database,
	)

	return url
}

