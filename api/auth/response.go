package auth

import (
	"userland/api/middleware"
)

type Response map[string] interface{}
type ErrorResponse struct {
	Code string
	Message string
}

type LoginResponse struct {
	RequireTfa bool `json:"require_tfa"`
	Token middleware.JWTToken `json:"access_token"`
}