package api

import (
	"github.com/rs/zerolog/log"
	"userland/store/postgres"
)

type Response map[string] interface{}
type ErrorResponse struct {
	Code string
	Message string
}


func GenerateErrorResponse(err error) ErrorResponse {
	var (
		errCode string
		errMsg string
	)

	switch err.(type) {
	case postgres.CustomErrorInterface:
		customError := err.(postgres.CustomErrorInterface)
		errMsg = customError.Error()
		errCode = customError.GetDatabaseErrorCode()
		log.Error().Stack().Err(customError.GetErr()).Msg("")
		break
	default:
		errCode = ErrInternalServerErrorCode
		errMsg = "internal server error"
		break
	}
	log.Err(err)
	return ErrorResponse {
		Code: errCode,
		Message: errMsg,
	}
}

