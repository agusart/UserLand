package postgres

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"time"
)

type ClientSession struct {
	Id uint
	Name string
	CreatedAt time.Time
}
type Session struct {
	Id uint `json:"-"`
	UserId uint `json:"-"`
	Client ClientSession
	IP string
	JwtId string `json:"-"`
	CreatedAt time.Time
	DeletedAt time.Time
	IsCurrent bool
}

type SessionStoreInterface interface {
	CreateNewSession(session Session) (*Session, error)
	UpdateSession(session Session) error
	DeleteSession(session Session) error
	GetSessionById(id uint) (*Session, error)
	GetSessionByUserId(userId uint) ([]Session, error)
	CreateClient(name string) (*ClientSession, error)
}


type SessionStore struct {
	db *sql.DB
}

func NewSessionStore(db *sql.DB) SessionStoreInterface {
	return SessionStore{db: db}
}

func (s SessionStore) CreateNewSession(session Session) (*Session, error) {
	insertSessionQuery := "insert into session(ip, jwt_id, user_id, client_id, created_at)" +
		"VALUES ($1, $2, $3, $4, $5) RETURNING id"

	session.CreatedAt = time.Now()
	var insertedId uint
	row, err := QueryRowPrepareStatement(
		s.db,
		insertSessionQuery,
		session.IP,
		session.JwtId,
		session.UserId,
		session.Client.Id,
		session.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	err = row.Scan(&insertedId)
	if err != nil {
		return nil, CustomError {
			ErrCantInsertUserSession,
			"cant create session",
			errors.Errorf("database error: %v", err),
		}
	}

	session.Id = insertedId
	return &session, nil
}

func (s SessionStore) UpdateSession(session Session) error {
	updateSessionSql := "UPDATE session SET ip=$1, jwt_id=$2, client_id=$3, user_id=$4, created_at=$5, deleted_at=$6 where id = $7"

	res, err := ExecPrepareStatement(
		s.db,
		updateSessionSql,
		session.IP,
		session.JwtId,
		session.Client.Id,
		session.UserId,
		session.CreatedAt,
		session.DeletedAt,
		session.Id,
	)
	if err != nil {
		return err
	}

	rowAffected, err := res.RowsAffected()
	if err != nil || rowAffected < 1 {
		return CustomError {
			ErrCantUpdateUserSession,
			"cant update session",
			errors.Errorf("database error: %v, row affected %d", err, rowAffected),
		}
	}

	return nil
}

func (s SessionStore) DeleteSession(session Session) error {
	session.DeletedAt = time.Now()
	return s.UpdateSession(session)
}

func (s SessionStore) GetSessionById(id uint) (*Session, error) {
	getSessionSql := "select s.id, s.user_id, s.ip, c.id, c.name, s.jwt_id, s.created_at, s.deleted_at from session s inner join client c on s.client_id = c.id where s.id = $1"
	row, err := QueryRowPrepareStatement(s.db, getSessionSql, id)
	if err != nil {
		return nil, err
	}

	return s.getSessionFromRow(row)
}

func (s SessionStore) GetSessionByUserId(userId uint) ([]Session, error) {
	getSessionSql := "select s.id, s.user_id, s.ip, c.id, c.name, s.jwt_id, s.created_at, s.deleted_at from session s inner join client c on s.client_id = c.id where s.user_id = $1"
	row, err := QueryPrepareStatement(s.db, getSessionSql, userId)
	if err != nil {
		return nil, err
	}

	return s.getSessionFromRows(row)
}

func (s SessionStore) CreateClient(name string) (*ClientSession, error) {
	insertClientQuery := "insert into client(name, created_at) values($1, $2) on conflict(name) do update set created_at = now() returning id"
	var insertedId uint
	createdAt := time.Now()
	row, err := QueryRowPrepareStatement(
		s.db,
		insertClientQuery,
		name,
		createdAt,
	)

	if err != nil {
		return nil, err
	}

	err = row.Scan(&insertedId)
	if err != nil {
		return nil, CustomError {
			ErrCantInsertUserSession,
			"failed create client",
			fmt.Errorf("failed to create client: %v", err),
		}
	}

	return &ClientSession{
		CreatedAt: createdAt,
		Id: insertedId,
		Name: name,

	}, nil
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
			&session.Client.Id,
			&session.Client.Name,
			&session.JwtId,
			&session.CreatedAt,
			&session.DeletedAt,
			)

		if err !=nil {
			return nil, CustomError{
				ErrGeneralDbErr,
				"internal server error",
				errors.Errorf("database error: %v", err),
			}
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
		&session.Client.Id,
		&session.Client.Name,
		&session.JwtId,
		&session.CreatedAt,
		&session.DeletedAt,
	)

	if err != nil {
		log.Print(err)
		if err == sql.ErrNoRows {
			return nil, CustomError {
				ErrUserNotfoundCode,
				"user session not found",
				errors.Errorf("database error: %v", err),
			}
		}

		return nil, CustomError {
			ErrGeneralDbErr,
			"internal server error",
			errors.Errorf("database error: %v", err),
		}
	}

	return &session, nil
}
