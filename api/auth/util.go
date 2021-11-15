package auth

import (
	"golang.org/x/crypto/bcrypt"
	"net/mail"
	"unicode"
	"userland/api"
	"userland/store/broker"
	"userland/store/postgres"
)

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func IsValidPassword(password string) bool {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	if len(password) >= 7 {
		hasMinLen = true
	}
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
}


func NewUserFromRegisterRequest(r RegisterRequest) postgres.User {
	user := postgres.User{}
	user.Password, _ = HashPassword(r.Password)
	user.Email = r.Email
	user.FullName = r.FullName

	return user
}


func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), api.HashPasswordCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func SendLog(msgBroker broker.MessageBrokerInterface, createdSeason postgres.Session) error {
	userLogJob := broker.UserLoginLogJob{
		LoggedInAt: createdSeason.CreatedAt,
		SessionId: createdSeason.Id,
		LoggedInIp: createdSeason.IP,
		UserId: createdSeason.UserId,
	}

	return msgBroker.SendLog(broker.MsgBrokerLogTopicName, userLogJob)
}

