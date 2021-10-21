package postgres

import (
	"database/sql"
	"time"
)

type Session struct {
	Id uint
	UserId uint
	IP string
	Name string
	JwtId string
	IsCurrent bool
	CreatedAt time.Time
	DeletedAt time.Time
}

type SessionStoreInterface interface {
	CreateNewSession(userId uint) (*Session, error)
	UpdateSession(session Session) error
	DeleteSession(session Session) error
	GetSessionById(id uint) (*Session, error)
	GetSessionByUserId(userId uint) (*Session, error)
}


type SessionStore struct {
	db *sql.DB
}

func NewSessionStore(db *sql.DB) SessionStoreInterface {
	return SessionStore{db: db}
}


func (s SessionStore) CreateNewSession(userId uint) (*Session, error) {
	panic("implement me")
}

func (s SessionStore) UpdateSession(session Session) error {
	panic("implement me")
}

func (s SessionStore) DeleteSession(session Session) error {
	panic("implement me")
}

func (s SessionStore) GetSessionById(id uint) (*Session, error) {
	panic("implement me")
}

func (s SessionStore) GetSessionByUserId(userId uint) (*Session, error) {
	panic("implement me")
}


