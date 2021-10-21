package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)


type JWTClaims struct {
	UserId uint
	SessionId uint
	TfaEnabled bool
	jwt.StandardClaims
}

type JWTToken struct {
	Token     string `json:"value"`
	ExpiredAt time.Time `json:"expired_at"`
	TokenType string
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
	GenerateAccessToken(userID, sessionId uint) (*JWTToken, error)
	GenerateRefreshToken(userID, sessionId uint) (*JWTToken, error)
	GenerateAccessTokenTfa(userID, sessionId uint) (*JWTToken, error)
	GenerateRefreshTokenTfa(userID, sessionId uint) (*JWTToken, error)
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

func (h JWTHandler) GenerateAccessToken(userId, sessionId uint) (*JWTToken, error) {
	claims := JWTClaims{
		UserId: userId,
		SessionId: sessionId,
	}
	return h.generateToken(claims, h.config.accesTokenExpiredTime)
}

func (h JWTHandler) GenerateAccessTokenTfa(userId, sessionId uint) (*JWTToken, error) {
	claims := JWTClaims{
		UserId: userId,
		SessionId: sessionId,
		TfaEnabled: true,
	}

	return h.generateToken(claims, h.config.accesTokenExpiredTime)
}

func (h JWTHandler) GenerateRefreshToken(userId, sessionId uint) (*JWTToken, error) {
	claims := JWTClaims{
		UserId: userId,
		SessionId: sessionId,
	}

	return h.generateToken(claims, h.config.accesTokenExpiredTime)
}

func (h JWTHandler) GenerateRefreshTokenTfa(userId, sessionId uint) (*JWTToken, error) {
	claims := JWTClaims{
		UserId: userId,
		SessionId: sessionId,
		TfaEnabled: true,
	}

	return h.generateToken(claims, h.config.accesTokenExpiredTime)
}

func (h JWTHandler) generateToken(claims JWTClaims, duration time.Duration) (*JWTToken, error) {
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