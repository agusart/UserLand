package postgres



const ErrUserAlreadyRegisteredCode = "ER-21"
const ErrCantInsertRegisterUser = "ER-22"
const ErrCantVerifyUser = "ER-23"

const ErrUserAlreadyVerifiedCode = "ER-31"

const ErrUserNotfoundCode = "ER-41"
const ErrGeneralDbErr = "ER-DB"



type CustomErrorInterface interface {
	Error() string
	GetDatabaseErrorCode() string
	GetErr() error
}

type CustomError struct {
	StatusCode string
	Err error
}

func (e CustomError) GetErr() error {
	return e.Err
}

func (e CustomError) GetDatabaseErrorCode() string {
	return e.StatusCode
}

func (e CustomError) Error() string {
	return e.Err.Error()
}






