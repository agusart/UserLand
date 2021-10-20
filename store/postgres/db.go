package postgres

import (
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
)

var db *sql.DB

type PGConfig struct {
	Host string
	Port int
	Username string
	Password string
	Database string
}

func NewPG(config PGConfig) (*sql.DB, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)
	connectionCfg, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config %w", err)
	}
	connStr := stdlib.RegisterConnConfig(connectionCfg)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return db, err
	}
	return db, nil
}






