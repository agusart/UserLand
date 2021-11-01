package postgres

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"time"
)

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


func prepareStatement(db *sql.DB, customSql string) (*sql.Stmt, error) {
	stmt, err := db.Prepare(customSql)
	if err != nil {
		return nil, CustomError {
			ErrGeneralDbErr,
			"internal sever error",
			errors.Errorf("database error: %v", err),
		}
	}

	return stmt, nil
}

func generateTfaCode() string {
	code := time.Now().UnixNano() % 1000000
	stringCode := fmt.Sprintf("%06d", code)

	return stringCode
}

func generateErr(err error) error {
	if err == sql.ErrNoRows {
		return CustomError{
			ErrUserNotfoundCode,
			"data not found",
			errors.Errorf("row not found : %v", err),
		}
	}

	return err
}
