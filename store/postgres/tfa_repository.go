package postgres

import "database/sql"

type TfaStoreInterface interface {
	CheckTfaBackupCode(userId uint, tfaCode string) (bool, error)
	DeleteTfaCode(userId uint, tfaCode string) error
	CreateTfaBackupCode(userId uint) ([]string, error)
}

type TfaStore struct {
	db *sql.DB
}

func (t TfaStore) CheckTfaBackupCode(userId uint, tfaCode string) (bool, error) {
	panic("implement me")
}

func (t TfaStore) DeleteTfaCode(userId uint, tfaCode string) error {
	panic("implement me")
}

func (t TfaStore) CreateTfaBackupCode(userId uint) ([]string, error) {
	panic("implement me")
}

func NewTfaStore(db *sql.DB) TfaStoreInterface {
	return TfaStore{db: db}
}
