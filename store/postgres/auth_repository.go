package postgres

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"time"
	"userland/store/redis"
)

type AuthStoreInterface interface {
	SendRegistrationVerificationCode(ctx context.Context, email string, duration time.Duration) (string, error)
	GetRegistrationCodeEmail(ctx context.Context, registrationToken string) (string, error)
	SendForgotPasswordVerificationCode(ctx context.Context, email string, duration time.Duration) (string, error)
	GetResetPasswordCodeEmail(ctx context.Context, resetPasswordToken string) (string, error)
}

type AuthStore struct {
	cache redis.CacheInterface
}



func NewAuthStore(db redis.CacheInterface) AuthStoreInterface {
	return AuthStore{cache: db}
}


func (a AuthStore) SendRegistrationVerificationCode(
	ctx context.Context,
	email string,
	duration time.Duration) (string, error) {
		token := "register:" + tokenGenerator()
		return a.insertToken(ctx, token, email, duration)
}

func (a AuthStore) GetRegistrationCodeEmail(ctx context.Context, registrationToken string) (string, error) {
	return a.getCodeEmail(ctx, "register:" + registrationToken)
}

func (a AuthStore) SendForgotPasswordVerificationCode(
	ctx context.Context,
	email string,
	duration time.Duration) (string, error) {
		token := "password:" + tokenGenerator()
		return a.insertToken(ctx, token, email, duration)
}

func (a AuthStore) GetResetPasswordCodeEmail(ctx context.Context, resetPasswordToken string) (string, error) {
	return a.getCodeEmail(ctx, "register:" + resetPasswordToken)
}


func (a AuthStore) insertToken(
	ctx context.Context,
	token,
	email string,
	duration time.Duration) (string, error){
	err := a.cache.SetWithTimout(ctx, token, email, duration)
	if err != nil {
		log.Print(err)
		return "", fmt.Errorf("cant send verification")
	}

	return token, nil
}



func (a AuthStore) getCodeEmail(ctx context.Context, token string) (string, error) {
	email, err := a.cache.Get(ctx, token)
	if err != nil {
		return "", err
	}

	return email, nil
}

