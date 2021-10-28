package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"time"
	"userland/api"
)


type JWTClaims struct {
	UserId      uint
	UserHasTfa  bool
	SessionId   uint
	TfaVerified bool
	jwt.StandardClaims
}



type JWTToken struct {
	Id string `json:"-"`
	Token     string `json:"value"`
	ExpiredAt time.Time `json:"expired_at"`
	TokenType string
}

type JWTConfig struct {
	jwtKey         []byte
	refreshTokenExpiredTime time.Duration
	accessTokenExpiredTime time.Duration
	alg jwt.SigningMethod
}

func NewJWTConfig(
	jwtKey string,
	refreshTokenExpiredTime,
	accessTokenExpiredTime time.Duration,
	alg jwt.SigningMethod) JWTConfig {
		return JWTConfig{
			jwtKey:         []byte(jwtKey),
			alg: alg,
			refreshTokenExpiredTime: refreshTokenExpiredTime,
			accessTokenExpiredTime: accessTokenExpiredTime,
		}
}

type JwtHandlerInterface interface {
	GenerateAccessToken(claim JWTClaims) (*JWTToken, error)
	GenerateRefreshToken(claim JWTClaims) (*JWTToken, error)
	ClaimToken(token string) (*JWTClaims, error)
}

type JWTHandler struct {
	config JWTConfig
}

func NewJWTHandler(config JWTConfig) JWTHandler {
	return JWTHandler{
		config: config,
	}
}

func (h JWTHandler) GenerateAccessToken(claim JWTClaims) (*JWTToken, error) {
	expiredTime := time.Now().Add(api.AccessTokenExpTime)
	claim.ExpiresAt = expiredTime.Unix()
	return h.generateToken(claim)
}

func (h JWTHandler) GenerateRefreshToken(claim JWTClaims) (*JWTToken, error) {
	expiredTime := time.Now().Add(api.RefreshTokenExpTime)
	claim.ExpiresAt = expiredTime.Unix()
	return h.generateToken(claim)
}


func (h JWTHandler) generateToken(claims JWTClaims) (*JWTToken, error) {
	token := jwt.NewWithClaims(
		h.config.alg,
		claims,
	)

	tokenString, err := token.SignedString(h.config.jwtKey)
	if err != nil {
		return nil, err
	}

	return &JWTToken{Token: tokenString, ExpiredAt: time.Unix(claims.ExpiresAt, 0), Id: claims.Id}, err
}

func (h JWTHandler) ClaimToken(token string) (*JWTClaims, error) {
	claims := &JWTClaims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return h.config.jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !tkn.Valid {
		return nil, err
	}

	return claims, err
}