package postgres

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"time"
)

func tokenGenerator() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}


func prepareStatement(db *sql.DB, customSql string) (*sql.Stmt, error) {
	stmt, err := db.Prepare(customSql)
	if err != nil {
		log.Print(err)
		return nil, CustomError {
			ErrGeneralDbErr,
			errors.New("database error"),
		}
	}

	return stmt, nil
}

func ExecPrepareStatement(db *sql.DB, customSql string, args ...interface{}) (sql.Result, error){
	stmt, err := prepareStatement(db, customSql)
	if err != nil {
		return nil, generateErr(err)
	}

	res, err := stmt.Exec(args...)
	if err != nil {
		return nil, generateErr(err)
	}

	return res, nil
}


func QueryRowPrepareStatement(db *sql.DB, customSql string, args ...interface{}) (*sql.Row, error){
	stmt, err := prepareStatement(db, customSql)
	if err != nil {
		return nil, generateErr(err)
	}

	row := stmt.QueryRow(args...)
	return row, nil
}


func QueryPrepareStatement(db *sql.DB, customSql string, args ...interface{}) (*sql.Rows, error){
	stmt, err := prepareStatement(db, customSql)
	if err != nil {
		return nil, generateErr(err)
	}
	row, err := stmt.Query(args...)
	return row, err
}


func generateTfaCode() string {
	code := time.Now().UnixNano() % 10000
	stringCode := fmt.Sprintf("%04d", code)

	return stringCode
}

func generateErr(err error) error {
	if err == sql.ErrNoRows {
		return CustomError{
			ErrUserNotfoundCode,
			fmt.Errorf("user not found"),
		}
	}

	return err
}
