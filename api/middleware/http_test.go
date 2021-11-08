package middleware

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"userland/store/redis"
	"userland/store/redis/mocks"
)

func TestRequiredJSONTypeFail(t *testing.T) {
	resource := "/test"
	Router.With(RequestMustBeJsonMiddleware).Post(resource, func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})


	response := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", resource, nil)

	Router.ServeHTTP(response, request)

	assert.Equal(t, response.Code, http.StatusBadRequest)
}


func TestRequiredJSONTypeSuccess(t *testing.T) {
	resource := "/test"
	Router.With(RequestMustBeJsonMiddleware).Post(resource, func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})


	response := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", resource, nil)
	request.Header.Set("Content-Type", "application/json")

	Router.ServeHTTP(response, request)

	assert.Equal(t, response.Code, http.StatusOK)
}


func TestRequireTokenAuthentication(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	cache := mocks.NewMockCacheInterface(mockCtrl)

	authMiddleware := NewAuthMiddleware(JwtHandler, cache)

	resource := "/test"
	Router.With(RequestMustBeJsonMiddleware, authMiddleware.UserAuthMiddleware).Post(resource, func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	response := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", resource, nil)
	request.Header.Set("Content-Type", "application/json")

	Router.ServeHTTP(response, request)

	assert.Equal(t, response.Code, http.StatusUnauthorized)
}

func TestInvalidToken(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	cache := mocks.NewMockCacheInterface(mockCtrl)

	authMiddleware := NewAuthMiddleware(JwtHandler, cache)

	resource := "/test"
	Router.With(RequestMustBeJsonMiddleware, authMiddleware.UserAuthMiddleware).Post(resource, func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	response := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", resource, nil)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer xyz")

	Router.ServeHTTP(response, request)

	assert.Equal(t, response.Code, http.StatusUnauthorized)
}


func TestRequireTokenAuthenticationSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	cache := mocks.NewMockCacheInterface(mockCtrl)

	authMiddleware := NewAuthMiddleware(JwtHandler, cache)

	resource := "/test"
	Router.With(RequestMustBeJsonMiddleware, authMiddleware.UserAuthMiddleware).Post(resource, func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	response := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", resource, nil)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+Token.Token)

	session := redis.SessionCache{JwtId: Claim.Id}

	cache.EXPECT().GetSessionCache(request.Context(), Claim.UserId, Claim.SessionId).Return(&session, nil)

	Router.ServeHTTP(response, request)

	assert.Equal(t, response.Code, http.StatusOK)
}



func TestRequireTfaFail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	cache := mocks.NewMockCacheInterface(mockCtrl)

	Claim.UserHasTfa = true
	Claim.TfaVerified = false

	tokenWithTfa, err := JwtHandler.GenerateAccessToken(Claim)
	assert.NoError(t, err)

	authMiddleware := NewAuthMiddleware(JwtHandler, cache)

	resource := "/test"
	Router.With(RequestMustBeJsonMiddleware, authMiddleware.UserAuthMiddleware, TfaRequiredMiddleware).Post(resource, func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	response := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", resource, nil)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+tokenWithTfa.Token)

	session := redis.SessionCache{JwtId: Claim.Id}

	cache.EXPECT().GetSessionCache(request.Context(), Claim.UserId, Claim.SessionId).Return(&session, nil)

	Router.ServeHTTP(response, request)

	assert.Equal(t, response.Code, http.StatusUnauthorized)
}


func TestRequireTfaSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	cache := mocks.NewMockCacheInterface(mockCtrl)

	Claim.UserHasTfa = true
	Claim.TfaVerified = true

	tokenWithTfa, err := JwtHandler.GenerateAccessToken(Claim)
	assert.NoError(t, err)

	authMiddleware := NewAuthMiddleware(JwtHandler, cache)

	resource := "/test"
	Router.With(RequestMustBeJsonMiddleware, authMiddleware.UserAuthMiddleware, TfaRequiredMiddleware).Post(resource, func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	response := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", resource, nil)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+tokenWithTfa.Token)

	session := redis.SessionCache{JwtId: Claim.Id}

	cache.EXPECT().GetSessionCache(request.Context(), Claim.UserId, Claim.SessionId).Return(&session, nil)

	Router.ServeHTTP(response, request)

	assert.Equal(t, response.Code, http.StatusOK)
}