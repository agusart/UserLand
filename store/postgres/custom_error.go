package postgres

import "github.com/rs/zerolog/log"


const ErrCantInsertUserSession = "ER-11"
const ErrCantUpdateUserSession = "ER-12"

const ErrUserAlreadyRegisteredCode = "ER-24"
const ErrCantInsertRegisterUser = "ER-22"
const ErrCantVerifyUser = "ER-23"
const ErrCantUpdateUser = "ER-25"
const ErrCantDeleteUser = "ER-25"
const ErrUserAlreadyVerifiedCode = "ER-31"
const ErrUserNotfoundCode = "ER-41"
const ErrGeneralDbErr = "ER-DB"



type CustomErrorInterface interface {
	Error() string
	GetMessage() string
	GetStatusCode() string
	PrintStackTrace()
}

type CustomError struct {
	StatusCode string
	Msg string
	Err error
}

func (e CustomError) PrintStackTrace() {
	log.Error().Stack().Err(e.Err).Msg("")
}

func (e CustomError) GetStatusCode() string {
	return e.StatusCode
}

func (e CustomError) GetMessage() string {
	return e.Msg
}


func (e CustomError) Error() string {
	return e.Err.Error()
}






