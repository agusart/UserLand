package postgres

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"time"
)

type UserLog struct {
	Id uint `json:"id"`
	UserId uint `json:"user_id"`
	SessionId uint `json:"session_name"`
	RemoteIp string `json:"remote_ip"`
	CreatedAt time.Time `json:"created_at"`
	DeletedAt time.Time `json:"deleted_at"`
}

type LogStoreInterface interface {
	 WriteUserLog(log UserLog) error
	 GetUserLogHistory(userId uint) ([]UserLog, error)
}

func NewLogStore(db *sql.DB) LogStoreInterface {
	return LogStore{
		db: db,
	}
}
type LogStore struct {
	db *sql.DB
}

func (l LogStore) WriteUserLog(userLog UserLog) error {
	insertLogQuery := "insert into audit_logs (user_id, session_id, remote_ip, created_at) values($1, $2, $3, $4)"
	userLog.CreatedAt = time.Now()

	var insertedId uint
	row, err := QueryRowPrepareStatement(
		l.db,
		insertLogQuery,
		userLog.UserId,
		userLog.SessionId,
		userLog.RemoteIp,
		time.Now(),
	)

	if err != nil {
		return err
	}

	err = row.Scan(&insertedId)
	if err != nil {
		return  CustomError {
			"",
			"cant insert log",
			errors.Errorf("database error: %v", err),
		}
	}

	return nil
}

func (l LogStore) GetUserLogHistory(userId uint) ([]UserLog, error) {
	getLogSql := "select * from audit_logs where user_id = $1"
	row, err := QueryPrepareStatement(l.db, getLogSql, userId)
	if err != nil {
		return nil, err
	}

	return l.getLogFromRows(row)
}

func (l LogStore) getLogFromRows(row *sql.Rows) ([]UserLog, error){
	if row == nil {
		return nil, fmt.Errorf("row from sql is nil")
	}

	var logUsers []UserLog
	for row.Next() {
		tmpLog := UserLog{}
		err := row.Scan(
			&tmpLog.Id,
			&tmpLog.UserId,
			&tmpLog.SessionId,
			&tmpLog.RemoteIp,
			&tmpLog.CreatedAt,
			&tmpLog.DeletedAt,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				return nil, CustomError {
					"",
					"log not found",
					errors.Errorf("database error: %v", err),
				}
			}

			return nil, CustomError {
				ErrGeneralDbErr,
				"internal server error",
				errors.Errorf("database error: %v", err),
			}

		}
		if tmpLog.DeletedAt.IsZero() {
			logUsers = append(logUsers, tmpLog)
		}
	}

	return logUsers, nil
}


