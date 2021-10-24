package postgres

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"time"
	"userland/store/redis"
)

type AuthStoreInterface interface {
	CreateRegistrationVerificationCode(ctx context.Context, email string, duration time.Duration) (string, error)
	GetRegistrationCodeEmail(ctx context.Context, registrationToken string) (string, error)
	CreateForgotPasswordVerificationCode(ctx context.Context, email string, duration time.Duration) (string, error)
	GetResetPasswordCodeEmail(ctx context.Context, resetPasswordToken string) (string, error)
	CreateTfaVerificationCode(ctx context.Context, email string, duration time.Duration) (string, error)
	GetTfaCodeEmail(ctx context.Context, tfaCode string) (string, error)
}

type AuthStore struct {
	cache redis.CacheInterface
}

func NewAuthStore(db redis.CacheInterface) AuthStoreInterface {
	return AuthStore{cache: db}
}

func (a AuthStore) CreateRegistrationVerificationCode(
	ctx context.Context,
	email string,
	duration time.Duration) (string, error) {
		token := tokenGenerator()
		return a.insertToken(ctx, "register:", token, email, duration)
}

func (a AuthStore) GetRegistrationCodeEmail(ctx context.Context, registrationToken string) (string, error) {
	return a.getCode(ctx, "register:" + registrationToken)
}

func (a AuthStore) CreateForgotPasswordVerificationCode(
	ctx context.Context,
	email string,
	duration time.Duration) (string, error) {
		token := tokenGenerator()
		return a.insertToken(ctx, "password:", token, email, duration)
}

func (a AuthStore) GetResetPasswordCodeEmail(ctx context.Context, resetPasswordToken string) (string, error) {
	return a.getCode(ctx, "password:" + resetPasswordToken)
}

func (a AuthStore) CreateTfaVerificationCode(ctx context.Context, email string, duration time.Duration) (string, error) {
	token := tokenGenerator()
	return a.insertToken(ctx,"tfa:", token, email, duration)
}

func (a AuthStore) GetTfaCodeEmail(ctx context.Context, tfaCode string) (string, error) {
	return a.getCode(ctx, "tfa:" + tfaCode)

}

func (a AuthStore) insertToken(
	ctx context.Context,
	prefix,
	token,
	email string,
	duration time.Duration) (string, error){
	err := a.cache.SetWithTimout(ctx, prefix+token, email, duration)
	if err != nil {
		return "", fmt.Errorf("cant send verification")
	}

	return token, nil
}



func (a AuthStore) getCode(ctx context.Context, token string) (string, error) {
	email, err := a.cache.Get(ctx, token)
	if err != nil {
		log.Err(err)
		return "", CustomError{
			StatusCode: ErrCantVerifyUser,
			Err: errors.New("cant verify User"),
		}
	}

	return email, nil
}

