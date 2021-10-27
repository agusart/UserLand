package api

import (
	"github.com/dgryski/dgoogauth"
	"github.com/rs/zerolog/log"
	"strings"
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
	log.Print(err)
	return ErrorResponse {
		Code: errCode,
		Message: errMsg,
	}
}


func VerifyTfaCode(secret, code string) (bool, error) {
	otpConfig := &dgoogauth.OTPConfig{
		Secret:      strings.TrimSpace(secret),
		WindowSize:  3,
		HotpCounter: 0,
	}

	trimmedToken := strings.TrimSpace(code)
	ok, err := otpConfig.Authenticate(trimmedToken)

	return ok, err
}

