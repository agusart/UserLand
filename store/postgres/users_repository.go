package postgres

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"time"
	"userland/store/redis"
)

type User struct {
	Id uint
	FullName string
	Email string
	Password string
	Verified bool
	TfaEnabled bool
	CreatedAt *time.Time
	DeletedAt *time.Time
	Location string
	Bio string
	Web string
	Picture string
}

type UserStoreInterface interface {
	RegisterUser(user User) error
	GetUserByEmail(email string) (*User, error)
	GetUserById(userId uint) (*User, error)
	IsUserVerified(email string) (bool, error)
	VerifyUser(email string) error
	UpdateUserPassword(userId uint, newPassword string) error
	UpdateUserBasicInfo(user User) error
	SaveUserTfaSecret(secret string, userId uint) error
	RemoveTfaStatus(userId uint) error
	DeleteUser(userId uint) error
}

type UserStore struct {
	db *sql.DB
	cache redis.CacheInterface
}

func (u UserStore) DeleteUser(userId uint) error {
	sqlDeleteUser := "update table users set deleted_at=now() where id=$1"
	res, err := ExecPrepareStatement(
		u.db,
		sqlDeleteUser,
		userId,
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
			StatusCode: ErrCantUpdateUser,
			Err: fmt.Errorf("cant update user"),
		}
	}

	return nil
}

func (u UserStore) RemoveTfaStatus(userId uint) error {
	sqlRemoveTfaStatus := "update table users set tfa_secret='', tfa_enabled=$1 where id=$2"
	res, err := ExecPrepareStatement(
		u.db,
		sqlRemoveTfaStatus,
		false,
		userId,
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
			StatusCode: ErrCantUpdateUser,
			Err: fmt.Errorf("cant update user"),
		}
	}

	return nil
}

func (u UserStore) SaveUserTfaSecret(secret string, userId uint) error {
	sqlUpdateUserSecret := "update table users set tfa_secret=$1, tfa_enabled=$2 where id=$3"
	res, err := ExecPrepareStatement(
		u.db,
		sqlUpdateUserSecret,
		secret,
		true,
		userId,
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
			StatusCode: ErrCantUpdateUser,
			Err: fmt.Errorf("cant update user"),
		}
	}

	return nil
}

func NewUserStore(db *sql.DB) UserStoreInterface {
	return UserStore{
		db: db,
	}
}


func (u UserStore) UpdateUserBasicInfo(user User) error {
	updateBasicInfoSql := "update table users set full_name = $1, location=$2, web=$3, bio=$4, email=$5 where id = $6"
	res, err := ExecPrepareStatement(
		u.db,
		updateBasicInfoSql,
		user.FullName,
		user.Location,
		user.Web,
		user.Bio,
		user.Email,
		user.Id,
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
			StatusCode: ErrCantUpdateUser,
			Err: fmt.Errorf("cant update user"),
		}
	}

	return nil
}



func (u UserStore) UpdateUserPassword(userId uint, newPassword string) error {
	updateSql := "UPDATE  users set password = $1 where id = $2"
	res, err := ExecPrepareStatement(u.db, updateSql, newPassword, userId)
	if err != nil {
		return errors.New(err.Error())
	}

	rowAffected, err := res.RowsAffected()
	if err != nil {
		return errors.New(err.Error())
	}

	if rowAffected < 1 {
		return CustomError{
			StatusCode: ErrCantVerifyUser,
			Err: errors.New("cant verify User"),
		}
	}

	return nil
}

func (u UserStore) VerifyUser(email string) error {
	updateSql := "UPDATE users set verified = true where email = $1"
	res, err := ExecPrepareStatement(u.db, updateSql, email)
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
			Err: errors.New("cant verify user"),
		}
	}

	return nil
}

func (u UserStore) RegisterUser(user User) error {
	existedUser, _ := u.GetUserByEmail(user.Email)
	if existedUser != nil  {
		return CustomError{ErrUserAlreadyRegisteredCode,
			errors.New("user already registered"),
		}
	}

	insertSql := "insert into users (full_name, password, email, created_at) values ($1, $2, $3, $4) RETURNING id"

	var insertedId int
	row, err := QueryRowPrepareStatement(u.db, insertSql, user.FullName, user.Password, user.Email, time.Now())
	if err != nil {
		return errors.New(err.Error())
	}

	err = row.Scan(&insertedId)
	if err != nil {
		return CustomError {
			ErrCantInsertRegisterUser,
			errors.New("failed to register"),
		}
	}

	return nil
}

func (u UserStore) GetUserByEmail(email string) (*User, error) {
	getUserSql := "select * from users where email = $1"
	row, err := QueryRowPrepareStatement(u.db, getUserSql, email)
	if err != nil {
		log.Err(err)
		return nil, CustomError {
			ErrGeneralDbErr,
			errors.New("database error"),
		}

	}

	return u.getUserFromRow(row)
}



func (u UserStore) GetUserById(userId uint) (*User, error) {
	getUserSql := "select * from users where id = $1"
	row, err := QueryRowPrepareStatement(u.db, getUserSql, userId)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	return u.getUserFromRow(row)
}

func (u UserStore) IsUserVerified(email string) (bool, error) {
	user, err := u.GetUserByEmail(email)
	if err != nil {
		return false, errors.New(err.Error())
	}

	if user == nil {
		return false, CustomError {
			ErrUserNotfoundCode,
			errors.New("user not found"),
		}
	}

	return user.Verified && user.DeletedAt == nil, nil
}

func (u UserStore) isUserExisted(email string) (bool, error) {
	sqlUserExistCheck := "select id, deleted_at from users where email = $1"

	var (
		id int
		deletedAt *time.Time
	)

	res, err := QueryRowPrepareStatement(u.db, sqlUserExistCheck, email)
	if err != nil {
		return false, errors.New(err.Error())
	}

	err = res.Scan(&id, &deletedAt)
	if err != nil {
		return false, CustomError {
			ErrGeneralDbErr,
			errors.New("database error"),
		}
	}

	return id != 0 && deletedAt == nil, nil
}

func (u UserStore) getUserFromRow(row *sql.Row) (*User, error){
	if row == nil {
		return nil, errors.New("row from sql is nil")
	}

	var user User
	err := row.Scan(
		&user.Id,
		&user.FullName,
		&user.Password,
		&user.Email,
		&user.Verified,
		&user.TfaEnabled,
		&user.CreatedAt,
		&user.DeletedAt,
		&user.Location,
		&user.Bio,
		&user.Web,
		&user.Picture,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, CustomError{
				ErrUserNotfoundCode,
				errors.New("user not found"),
			}
		}

		return nil, CustomError {
			ErrGeneralDbErr,
			errors.New("database error"),
		}
	}

	return &user, nil
}
