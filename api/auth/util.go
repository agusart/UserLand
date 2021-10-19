package auth

import (
	"golang.org/x/crypto/bcrypt"
	"net/mail"
	"unicode"
	"userland/api"
	"userland/store/postgres"
)

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func isValidPassword(password string) bool {
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


func GenerateErrorResponse(err error) ErrorResponse {
	var (
		errCode string
		errMsg string
	)

	switch err.(type) {
	case ErrUserAlreadyRegistered:
		errCode = api.ErrUserAlreadyRegistered
		errMsg = err.Error()
		break
	case error:
		errCode = api.ErrInternalServerError
		errMsg = "internal server error"
		break
	}

	return ErrorResponse{
		Code: errCode,
		Message: errMsg,
	}
}

