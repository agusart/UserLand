package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)


type JWTClaims struct {
	Username string
	SessionId int
	jwt.StandardClaims
}

type JWTToken struct {
	Token     string
	ExpiredAt time.Time
}

type JWTConfig struct {
	jwtKey         []byte
	refreshTokenExpiredTime time.Duration
	accesTokenExpiredTime time.Duration
	alg jwt.SigningMethod
}

func NewJWTConfig(
	jwtKey string,
	refreshTokenExpiredTime,
	accesTokenExpiredTime time.Duration,
	alg jwt.SigningMethod) JWTConfig {
		return JWTConfig{
			jwtKey:         []byte(jwtKey),
			alg: alg,
			refreshTokenExpiredTime: refreshTokenExpiredTime,
			accesTokenExpiredTime: accesTokenExpiredTime,
		}
}

type JwtHandlerInterface interface {
	GenerateAccesToken(username string, sessionId int) (*JWTToken, error)
	GenerateRefreshToken(username string, sessionId int) (*JWTToken, error)
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

func (h JWTHandler) GenerateAccesToken(username string, sessionId int) (*JWTToken, error) {
	return h.generateToken(username, sessionId, h.config.accesTokenExpiredTime)
}

func (h JWTHandler) GenerateRefreshToken(username string, sessionId int) (*JWTToken, error) {
	return h.generateToken(username, sessionId, h.config.refreshTokenExpiredTime)
}

func (h JWTHandler) generateToken(username string, sessionId int, duration time.Duration) (*JWTToken, error) {

	claims := JWTClaims{
		Username: username,
		SessionId: sessionId,
	}

	expiredTime := time.Now().Add(duration)
	claims.ExpiresAt = expiredTime.Unix()

	token := jwt.NewWithClaims(
		h.config.alg,
		claims,
	)

	tokenString, err := token.SignedString(h.config.jwtKey)
	if err != nil {
		return nil, err
	}

	return &JWTToken{Token: tokenString, ExpiredAt: expiredTime}, err
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