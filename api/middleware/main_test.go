package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
	"log"
	"testing"
	"userland/api"
)

var (
	JwtHandler JwtHandlerInterface
	Token      JWTToken
	Router     *chi.Mux
	Claim JWTClaims
)

func TestMain(m *testing.M){
	setUp()
	m.Run()
}


func setUp() {

	jwtConfig := NewJWTConfig(
		"asdf",
		api.RefreshTokenExpTime,
		api.AccessTokenExpTime,
		jwt.SigningMethodHS256,
	)
	JwtHandler = NewJWTHandler(jwtConfig)

	Claim = JWTClaims{
		UserId: 1,
		SessionId: 1,
	}

	Claim.Id = "id"
	jwtToken, err := JwtHandler.GenerateAccessToken(Claim)

	if err != nil {
		log.Fatalf("err: %v", err)
	}
	if jwtToken == nil {
		log.Fatal("Token nil")
	}

	Token = *jwtToken

	Router = chi.NewRouter()
}
