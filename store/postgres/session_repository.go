package postgres

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"time"
)

type Session struct {
	Id uint
	UserId uint
	IP string
	Name string
	JwtId string
	CreatedAt time.Time
	DeletedAt time.Time
}

type SessionStoreInterface interface {
	CreateNewSession(session Session) (*Session, error)
	UpdateSession(session Session) error
	DeleteSession(session Session) error
	GetSessionById(id uint) (*Session, error)
	GetSessionByUserId(userId uint) ([]Session, error)
}


type SessionStore struct {
	db *sql.DB
}


func NewSessionStore(db *sql.DB) SessionStoreInterface {
	return SessionStore{db: db}
}


func (s SessionStore) CreateNewSession(session Session) (*Session, error) {
	insertSessionQuery := "insert into session(ip, jwt_id, user_id, name, created_at)" +
		"VALUES ($1, $2, $3, $4, $5) RETURNING id"

	session.CreatedAt = time.Now()
	var insertedId uint
	row, err := QueryRowPrepareStatement(
		s.db,
		insertSessionQuery,
		session.IP,
		session.JwtId,
		session.UserId,
		session.Name,
		session.CreatedAt,
		)

	if err != nil {
		return nil, err
	}

	err = row.Scan(&insertedId)
	if err != nil {
		return nil, CustomError {
			ErrCantInsertRegisterUser,
			fmt.Errorf("failed to register"),
		}
	}

	session.Id = insertedId
	return &session, nil
}

func (s SessionStore) UpdateSession(session Session) error {
	updateSessionSql := "UPDATE session SET ip=$1, jwt_id=$2, name=$3, user_id=$4, created_at=$5, deleted_at=$6 where id = $7"

	res, err := ExecPrepareStatement(
		s.db,
		updateSessionSql,
		session.IP,
		session.JwtId,
		session.Name,
		session.UserId,
		session.CreatedAt,
		session.DeletedAt,
		session.Id,
		)
	if err != nil {
		return err
	}

	rowAffected, err := res.RowsAffected()
	if err != nil {
		return CustomError {
			ErrGeneralDbErr,
			errors.New("database error"),
		}
	}

	if rowAffected < 1 {
		return CustomError{
			StatusCode: ErrCantVerifyUser,
			Err: fmt.Errorf("cant update session"),
		}
	}

	return nil
}

func (s SessionStore) DeleteSession(session Session) error {
	session.DeletedAt = time.Now()
	return s.UpdateSession(session)
}

func (s SessionStore) GetSessionById(id uint) (*Session, error) {
	getSessionSql := "select * from session where id = $1"
	row, err := QueryRowPrepareStatement(s.db, getSessionSql, id)
	if err != nil {
		return nil, err
	}

	return s.getSessionFromRow(row)
}

func (s SessionStore) GetSessionByUserId(userId uint) ([]Session, error) {
	getSessionSql := "select * from session where user_id = $1"
	row, err := QueryPrepareStatement(s.db, getSessionSql, userId)
	if err != nil {
		return nil, err
	}

	return s.getSessionFromRows(row)
}

func (s SessionStore) getSessionFromRows(rows *sql.Rows) ([]Session, error){
	defer rows.Close()
	var sessions []Session
	for rows.Next() {
		session := Session{}
		err := rows.Scan(
			&session.Id,
			&session.UserId,
			&session.IP,
			&session.Name,
			&session.JwtId,
			&session.CreatedAt,
			&session.DeletedAt,
			)

		if err !=nil {
			return nil, err
		}
		if session.DeletedAt.IsZero() {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

func (s SessionStore) getSessionFromRow(row *sql.Row) (*Session, error){
	if row == nil {
		return nil, fmt.Errorf("row from sql is nil")
	}


	var session Session
	err := row.Scan(
		&session.Id,
		&session.UserId,
		&session.IP,
		&session.Name,
		&session.JwtId,
		&session.CreatedAt,
		&session.DeletedAt,
	)

	if err != nil {
		log.Print(err)
		if err == sql.ErrNoRows {
			return nil, CustomError{
				ErrUserNotfoundCode,
				fmt.Errorf("user not found"),
			}
		}

		return nil, CustomError {
			ErrGeneralDbErr,
			errors.New("database error"),
		}
	}

	return &session, nil
}
