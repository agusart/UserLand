package postgres

import "fmt"



const ErrUserAlreadyRegisteredCode = "ER-21"
const ErrCantInsertRegisterUser = "ER-21"

const ErrUserAlreadyVerifiedCode = "ER-31"

const ErrUserNotfoundCode = "ER-41"



var GeneralDatabaseErr = CustomError {
	"ER-DB",
	fmt.Errorf("database error"),
}


type CustomErrorInterface interface {
	Error() string
	GetErrorCode() string
}

type CustomError struct {
	StatusCode string
	Err error
}


func (e CustomError) GetErrorCode() string {
	return e.StatusCode
}

func (e CustomError) Error() string {
	return e.Err.Error()
}






