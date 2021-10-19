package postgres

import (
	"context"
	"database/sql"
	"time"
	"userland/store/redis"
)

type User struct {
	Id uint
	FullName string
	Email string
	Password string
	Verified bool
}

type UserStoreInterface interface {
	RegisterUser(user User) error
	GetUserByEmail(email string) (*User, error)
	IsUserVerified(email string) (bool, error)
	SendVerification(ctx context.Context, email string, duration time.Duration) (string, error)

}

type UserStore struct {
	db *sql.DB
	cache redis.CacheInterface
}

func NewUserStore(db *sql.DB, cache redis.CacheInterface) UserStoreInterface {
	return UserStore{
		db: db,
		cache: cache,
	}
}

func (u UserStore) RegisterUser(user User) error {
	stmt, err := u.db.Prepare(
		"insert into users (full_name, password, email) values ($1, $2, $3)",
	)

	if err != nil {
		return err
	}

	res, err := stmt.Exec(user.FullName, user.Password, user.Email)
	if err != nil{
		return err
	}

	_, err = res.LastInsertId()
	if err != nil{
		return err
	}

	//token := tokenGenerator()
	//err = u.cache.SetWithTimout(ctx, token, strconv.FormatInt(id, 10), 15 * time.Minute)
	//if err != nil{
	//	return  err
	//}
	//
	//api.SendEmail(user.Email, token)
	return nil
}

func (u UserStore) GetUserByEmail(email string) (*User, error) {
	stmt, err := u.db.Prepare("select * from users where email = $1")
	if err != nil {
		return nil, err
	}

	var user User
	row := stmt.QueryRow(email)
	if err := row.Scan(&user.Id, &user.FullName, &user.Password, &user.Email); err != nil {
		return nil, err
	}

	return &user, nil
}

func (u UserStore) IsUserVerified(email string) (bool, error) {
	user, err := u.GetUserByEmail(email)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, nil
	}

	return !user.Verified, nil
}

func (u UserStore) SendVerification(ctx context.Context, email string, duration time.Duration) (string, error) {
	token := tokenGenerator()
	err := u.cache.SetWithTimout(ctx, email, token, duration)
	if err != nil {
		return "", nil
	}
	return token, nil
}

func (u UserStore) isUserExisted(email string) bool {
	stmt, _ := u.db.Prepare("select id from users where email = ?")
	res, _ := stmt.Exec(email)

	affectedRow, _ := res.RowsAffected()
	return affectedRow != 0
}


