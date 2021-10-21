package postgres

import (
	"database/sql"
	"fmt"
	"log"
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
}

type UserStoreInterface interface {
	RegisterUser(user User) error
	GetUserByEmail(email string) (*User, error)
	GetUserById(userId uint) (*User, error)
	IsUserVerified(email string) (bool, error)
	VerifyUser(email string) error
	UpdateUserPassword(userId uint, newPassword string) error
}

type UserStore struct {
	db *sql.DB
	cache redis.CacheInterface
}

func (u UserStore) UpdateUserPassword(userId uint, newPassword string) error {
	updateSql := "UPDATE  users set password = $1 where id = $2"
	res, err := u.execPrepareStatement(updateSql, []interface{}{newPassword, userId})
	if err != nil {
		return err
	}

	rowAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowAffected < 1 {
		return CustomError{
			StatusCode: ErrCantVerifyUser,
			Err: fmt.Errorf("cant verify User"),
		}
	}

	return nil
}
func (u UserStore) VerifyUser(email string) error {
	updateSql := "UPDATE users set verified = true where email = $1"
	res, err := u.execPrepareStatement(updateSql, []interface{}{email})
	if err != nil {
		return err
	}

	rowAffected, err := res.RowsAffected()
	if err != nil {
		return GeneralDatabaseErr
	}

	if rowAffected < 1 {
		return CustomError{
			StatusCode: ErrCantVerifyUser,
			Err: fmt.Errorf("cant verify user"),
		}
	}

	return nil
}

func NewUserStore(db *sql.DB) UserStoreInterface {
	return UserStore{
		db: db,
	}
}

func (u UserStore) RegisterUser(user User) error {
	existedUser, err := u.GetUserByEmail(user.Email)
	if err != nil {
		return GeneralDatabaseErr
	}

	if existedUser != nil  {
		return CustomError{ErrUserAlreadyRegisteredCode,
			fmt.Errorf("user already registered"),
		}
	}

	stmt, err := u.db.Prepare(
		"insert into users (full_name, password, email, created_at) values ($1, $2, $3, $4) RETURNING id",
	)
	if err != nil {
		return GeneralDatabaseErr
	}

	var insertedId int
	err = stmt.QueryRow(user.FullName, user.Password, user.Email, time.Now()).Scan(&insertedId)
	if err != nil {
		return CustomError {
			ErrCantInsertRegisterUser,
			fmt.Errorf("failed to register"),
		}
	}

	return nil
}

func (u UserStore) GetUserByEmail(email string) (*User, error) {
	stmt, err := u.db.Prepare("select * from users where email = $1")
	if err != nil {
		return nil, GeneralDatabaseErr
	}

	var user User
	row := stmt.QueryRow(email)

	err = row.Scan(
		&user.Id,
		&user.FullName,
		&user.Password,
		&user.Email,
		&user.Verified,
		&user.TfaEnabled,
		&user.CreatedAt,
		&user.DeletedAt,
		)

	if err != nil {
		log.Println(err)
		if err == sql.ErrNoRows {
			return nil, CustomError{
				ErrUserNotfoundCode,
				fmt.Errorf("user not found"),
			}
		}

		return nil, GeneralDatabaseErr
	}

	return &user, nil
}


func (u UserStore) GetUserById(userId uint) (*User, error) {
	stmt, err := u.db.Prepare("select * from users where id = $1")
	if err != nil {
		return nil, GeneralDatabaseErr
	}

	var user User
	row := stmt.QueryRow(userId)

	err = row.Scan(
		&user.Id,
		&user.FullName,
		&user.Password,
		&user.Email,
		&user.Verified,
		&user.TfaEnabled,
		&user.CreatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		log.Println(err)
		if err == sql.ErrNoRows {
			return nil, CustomError{
				ErrUserNotfoundCode,
				fmt.Errorf("user not found"),
			}
		}

		return nil, GeneralDatabaseErr
	}

	return &user, nil
}

func (u UserStore) IsUserVerified(email string) (bool, error) {
	user, err := u.GetUserByEmail(email)
	if err != nil {
		return false, err
	}

	if user == nil {
		return false, CustomError {
			ErrUserNotfoundCode,
			fmt.Errorf("user not found"),
		}
	}

	return user.Verified && user.DeletedAt == nil, nil
}

func (u UserStore) isUserExisted(email string) (bool, error) {
	stmt, err := u.db.Prepare("select id, deleted_at from users where email = $1")
	if err != nil {
		return false, GeneralDatabaseErr
	}

	var (
		id int
		deletedAt *time.Time
	)
	err = stmt.QueryRow(email).Scan(&id, &deletedAt)
	if err != nil {
		return false, GeneralDatabaseErr
	}

	return id != 0 && deletedAt == nil, nil
}

func (u UserStore) execPrepareStatement(customSql string, args ...interface{}) (sql.Result, error){
	stmt, err := u.db.Prepare(customSql)
	if err != nil {
		return nil, GeneralDatabaseErr
	}

	res, err := stmt.Exec(args)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, CustomError{
				ErrUserNotfoundCode,
				fmt.Errorf("user not found"),
			}
		}

		return nil, GeneralDatabaseErr
	}

	return res, nil
}

