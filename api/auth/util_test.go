package auth

import (
	"fmt"
	"log"
	"testing"
	"userland/store/postgres"
)

func TestGenerateErrorResponse(t *testing.T) {
	err := postgres.CustomError{
		"asd",fmt.Errorf("asdfghjk"),
	}

	res := GenerateErrorResponse(err)
	log.Print(res)
}
