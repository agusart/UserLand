package postgres

import (
	"database/sql"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"time"
)

type ImgInfo struct {
	FileName string
	OwnerId uint
}

type User struct {
	Id uint
	FullName string
	Email string
	Password string
	Verified bool
	TfaEnabled bool
	CreatedAt *time.Time
	DeletedAt *time.Time
	Location sql.NullString
	Bio sql.NullString
	Web sql.NullString
	Picture sql.NullString
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
	DeleteUser(userId uint) error
	ChangeUserEmail(userId uint, email string) error
	SavePasswordToHistory(userId uint, passwordHash string) error
	GetPasswordHistory(userId uint) ([]string, error)
	SaveImage(imgInfo ImgInfo) error
	DeleteImage(userId uint) error
}

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) UserStoreInterface {
	return UserStore{
		db: db,
	}
}

func (u UserStore) GetUserByEmail(email string) (*User, error) {
	getUserSql := "select * from users where email = $1"
	row, err := QueryRowPrepareStatement(u.db, getUserSql, email)
	if err != nil {
		log.Err(err)
		return nil, CustomError {
			ErrUserNotfoundCode,
			"user not found",
			errors.Errorf("database error: %v", err),
		}

	}

	return u.getUserFromRow(row)
}

func (u UserStore) GetUserById(userId uint) (*User, error) {
	getUserSql := "select * from users where id = $1"
	row, err := QueryRowPrepareStatement(u.db, getUserSql, userId)
	if err != nil {
		return nil, err
	}

	return u.getUserFromRow(row)
}

func (u UserStore) IsUserVerified(email string) (bool, error) {
	user, err := u.GetUserByEmail(email)
	if err != nil {
		return false, err
	}

	if user == nil {
		return false, CustomError {
			ErrUserNotfoundCode,
			"user not found",
			errors.Errorf("database error: %v", err),
		}
	}

	return user.Verified && user.DeletedAt == nil, nil
}

func (u UserStore) VerifyUser(email string) error {
	updateSql := "UPDATE users set verified = true where email = $1"
	res, err := ExecPrepareStatement(u.db, updateSql, email)
	if err != nil {
		return err
	}

	rowAffected, err := res.RowsAffected()
	if err != nil || rowAffected < 1 {
		return CustomError {
			ErrCantVerifyUser,
			"Cant verify user",
			errors.Errorf("database error: %v, row affected: %d", err, rowAffected),
		}
	}

	return nil
}

func (u UserStore) UpdateUserPassword(userId uint, newPassword string) error {
	updatePasswordSql := "UPDATE  users set password = $1 where id = $2"
	res, err := ExecPrepareStatement(u.db, updatePasswordSql, newPassword, userId)
	if err != nil {
		return err
	}

	rowAffected, err := res.RowsAffected()
	if err != nil || rowAffected < 1 {
		return CustomError{
			ErrCantUpdateUser,
			"cant update password",
			errors.Errorf("error : %v, row affected: %d", err, rowAffected),
		}
	}

	return u.SavePasswordToHistory(userId, newPassword)
}

func (u UserStore) UpdateUserBasicInfo(user User) error {
	updateBasicInfoSql := "update users set full_name=$1, location=$2, web=$3, bio=$4, email=$5, picture=$6 where id=$7"
	res, err := ExecPrepareStatement(
		u.db,
		updateBasicInfoSql,
		user.FullName,
		user.Location.String,
		user.Web.String,
		user.Bio.String,
		user.Email,
		user.Picture.String,
		user.Id,
	)
	if err != nil {
		return err
	}

	rowAffected, err := res.RowsAffected()
	if err != nil || rowAffected < 1 {
		return CustomError {
			ErrCantUpdateUser,
			"cant update user info",
			errors.Errorf("database error: %v, row affected: %d", err, rowAffected),
		}
	}

	return nil
}

func (u UserStore) SaveUserTfaSecret(secret string, userId uint) error {
	sqlUpdateUserSecret := "update users set tfa_secret=$1, tfa_enabled=$2 where id=$3"
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
	if err != nil || rowAffected < 1{
		return CustomError {
			ErrCantUpdateUser,
			"cant activate tfa",
			errors.Errorf("database error: %v, row affected: %d", err, rowAffected),
		}
	}

	return nil
}

func (u UserStore) DeleteUser(userId uint) error {
	sqlDeleteUser := "update users set deleted_at=now() where id=$1"
	res, err := ExecPrepareStatement(
		u.db,
		sqlDeleteUser,
		userId,
	)

	if err != nil {
		return err
	}

	rowAffected, err := res.RowsAffected()
	if err != nil || rowAffected < 1 {
		return CustomError {
			ErrCantDeleteUser,
			"Error when try deleting account",
			errors.Errorf("database error: %v, row affected: %d", err, rowAffected),
		}
	}

	return nil
}

func (u UserStore) ChangeUserEmail(userId uint, email string) error {
	sqlUpdateEmail := "update users set email=$1 where id=$2"
	res, err := ExecPrepareStatement(
		u.db,
		sqlUpdateEmail,
		email,
		userId,
	)

	if err != nil {
		return err
	}

	rowAffected, err := res.RowsAffected()
	if err != nil || rowAffected < 1 {
		return CustomError {
			ErrCantUpdateUser,
			"error when try change user email",
			errors.Errorf("database error: %v, row affected: %d", err, rowAffected),
		}
	}

	return nil
}

func (u UserStore) SavePasswordToHistory(userId uint, passwordHash string) error {
	savePasswordSql := "insert into password_history(user_id, password, created_at) values ($1, $2, now()) returning id"

	var insertedId int
	row, err := QueryRowPrepareStatement(u.db, savePasswordSql, userId, passwordHash)
	if err != nil {
		return err
	}

	err = row.Scan(&insertedId)
	if err != nil {
		return CustomError {
			ErrCantInsertRegisterUser,
			"failed to register",
			err,
		}
	}

	return nil
}


func (u UserStore) GetPasswordHistory(userId uint) ([]string, error) {
	sqlGetPasswordHistory := "select password from password_history where id = $1 order by created_at desc limit 3"
	rows, err := QueryPrepareStatement(u.db, sqlGetPasswordHistory, userId)
	if err != nil {
		return nil,  err
	}

	var res []string

	for rows.Next() {
		var passTmp string

		err := rows.Scan(&passTmp)
		if err != nil {
			return nil, errors.Errorf("cant scan history password: %v", err)
		}

		res = append(res, passTmp)
	}

	return res, nil
}

func (u UserStore) SaveImage(imgInfo ImgInfo) error {
	user, err := u.GetUserById(imgInfo.OwnerId)
	if err != nil {
		return err
	}

	user.Picture = sql.NullString{String: imgInfo.FileName, Valid: true}
	return u.UpdateUserBasicInfo(*user)
}

func (u UserStore) DeleteImage(userId uint) error {
	user, err := u.GetUserById(userId)
	if err != nil {
		return err
	}

	user.Picture = sql.NullString{}
	return u.UpdateUserBasicInfo(*user)
}

func (u UserStore) RegisterUser(user User) error {
	existedUser, err := u.GetUserByEmail(user.Email)
	if err != nil {
		return err
	}

	if existedUser != nil  {
		err := errors.New("user already registered")
		return CustomError{ErrUserAlreadyRegisteredCode,
			err.Error(),
			err,
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
			"error when register user",
			errors.Errorf("database error: %d", err),
		}
	}

	return nil
}

func (u UserStore) isUserExisted(email string) (bool, error) {
	sqlUserExistCheck := "select id, deleted_at from users where email = $1"

	var (
		id int
		deletedAt *time.Time
	)

	res, err := QueryRowPrepareStatement(u.db, sqlUserExistCheck, email)
	if err != nil {
		return false, err
	}

	err = res.Scan(&id, &deletedAt)
	if err != nil {
		return false, CustomError {
			ErrGeneralDbErr,
			"internal server error",
			errors.Errorf("database error: %v", err),
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
				"user not found",
				errors.Errorf("database error: %v", err),
			}
		}

		return nil, CustomError {
			ErrGeneralDbErr,
			"internal server error",
			errors.Errorf("database error: %v", err.Error()),
		}
	}

	return &user, nil
}