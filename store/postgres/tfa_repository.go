package postgres

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"time"
)

type TfaStoreInterface interface {
	CheckTfaBackupCode(userId uint, tfaCode string) (bool, error)
	DeleteTfaCode(userId uint, tfaCode string) error
	CreateTfaBackupCode(userId uint) ([]string, error)
}

type TfaStore struct {
	db *sql.DB
}

func (t TfaStore) CheckTfaBackupCode(userId uint, tfaCode string) (bool, error) {
	sqlStatement := "select id, deleted_at from tfa_backup_code where user_id = $1 and code = $2"

	var (
		id int
		deletedAt *time.Time
	)

	res, err := QueryRowPrepareStatement(t.db, sqlStatement, userId, tfaCode)
	if err != nil {
		return false, err
	}

	err = res.Scan(&id, &deletedAt)
	if err != nil {
		return false, CustomError {
			ErrGeneralDbErr,
			errors.New("database error"),
		}
	}

	return id != 0 && deletedAt !=nil , nil
}

func (t TfaStore) DeleteTfaCode(userId uint, tfaCode string) error {
	sqlDeleteStatement := "update tfa_backup_code set deleted_at = now() where user_id = $1 and code = $2"
	res, err := ExecPrepareStatement(t.db, sqlDeleteStatement, userId, tfaCode)

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
		return fmt.Errorf("cant delete tfa code")
	}

	return nil
}

func (t TfaStore) CreateTfaBackupCode(userId uint) ([]string, error) {
	var backupCodes []string
	sqlInsertTfaCode := "insert into tfa_backup_code(user_id, code, created_at) values($1, $2, $3) returning id"

	for {
		tfaCode := generateTfaCode()
		res, err := QueryRowPrepareStatement(t.db, sqlInsertTfaCode, tfaCode, time.Now())
		if err != nil {
			return nil, err
		}

		var tfaInsertedId uint
		err = res.Scan(&tfaInsertedId)
		if err != nil {
			return nil, err
		}

		if tfaInsertedId == 0 {
			continue
		}

		backupCodes = append(backupCodes, tfaCode)
		if len(backupCodes) > 5 {
			break
		}
	}

	return backupCodes, nil
}

func NewTfaStore(db *sql.DB) TfaStoreInterface {
	return TfaStore{db: db}
}
