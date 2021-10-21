package postgres

import "fmt"



const ErrUserAlreadyRegisteredCode = "ER-21"
const ErrCantInsertRegisterUser = "ER-22"
const ErrCantVerifyUser = "ER-23"

const ErrUserAlreadyVerifiedCode = "ER-31"

const ErrUserNotfoundCode = "ER-41"



var GeneralDatabaseErr = CustomError {
	"ER-DB",
	fmt.Errorf("database error"),
}


type CustomErrorInterface interface {
	Error() string
	GetDatabaseErrorCode() string
}

type CustomError struct {
	StatusCode string
	Err error
}


func (e CustomError) GetDatabaseErrorCode() string {
	return e.StatusCode
}

func (e CustomError) Error() string {
	return e.Err.Error()
}






