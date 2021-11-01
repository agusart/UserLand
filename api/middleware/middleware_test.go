package middleware

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"userland/api"
	"userland/store/redis"
	"userland/store/redis/mocks"
)

type handler struct {}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestRequestMustBeJsonMiddlewareSuccest(t *testing.T) {
	middleware := RequestMustBeJsonMiddleware(handler{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Set("Content-Type", "application/json")

	middleware.ServeHTTP(response, request)
	assert.Equal(t, response.Code, http.StatusOK)
}


func TestRequestMustBeJsonMiddlewareBadRequest(t *testing.T) {
	middleware := RequestMustBeJsonMiddleware(handler{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)

	middleware.ServeHTTP(response, request)
	assert.Equal(t, response.Code, http.StatusBadRequest)
}


func TestTfaRequiredMiddleware(t *testing.T) {
	middleware := TfaRequiredMiddleware(handler{})
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)

	claim := JWTClaims{
		TfaVerified: true,
	}

	customCtx := request.Context()
	customCtx = context.WithValue(customCtx, api.ContextClaimsJwt, claim)

	middleware.ServeHTTP(response, request.WithContext(customCtx))
	assert.Equal(t, http.StatusOK, response.Code)
}


func TestTfaRequiredMiddlewareFail(t *testing.T) {
	middleware := TfaRequiredMiddleware(handler{})
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)

	claim := JWTClaims{
		TfaVerified: false,
	}

	customCtx := request.Context()
	customCtx = context.WithValue(customCtx, api.ContextClaimsJwt, claim)

	middleware.ServeHTTP(response, request.WithContext(customCtx))
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestAuthMiddlewareSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Set("Authorization", "Bearer "+Token.Token)

	session := redis.SessionCache{JwtId: Claim.Id}
	cache := mocks.NewMockCacheInterface(mockCtrl)

	cache.EXPECT().GetSessionCache(context.Background(), Claim.UserId, Claim.SessionId).Return(&session, nil)

	authMiddleware := NewAuthMiddleware(JwtHandler, cache)

	middleware := authMiddleware.UserAuthMiddleware(handler{})
	middleware.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
}