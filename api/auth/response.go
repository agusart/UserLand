package auth

import (
	"userland/api/middleware"
)


type LoginResponse struct {
	RequireTfa bool `json:"require_tfa"`
	Token middleware.JWTToken `json:"access_token"`
}