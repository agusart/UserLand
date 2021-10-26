package session

import (
	"encoding/json"
	"net/http"
	"time"
	"userland/api"
	"userland/api/middleware"
	"userland/store/postgres"
	"userland/store/redis"
)

func ListSession(sessionStore postgres.SessionStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claim :=  r.Context().Value(api.ContextClaimsJwt).(middleware.JWTClaims)
		sessions, err := sessionStore.GetSessionByUserId(claim.UserId)
		for _, s := range sessions{
			s.IsCurrent = s.JwtId == claim.Id
		}

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ListSessionResponse{
			Session: sessions,
		})
	}
}

func EndSession(
		sessionStore postgres.SessionStoreInterface,
		cache redis.CacheInterface,
	) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claim :=  r.Context().Value(api.ContextClaimsJwt).(middleware.JWTClaims)
		currentSession, err := sessionStore.GetSessionById(claim.SessionId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		err = sessionStore.DeleteSession(*currentSession)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		err = cache.DeleteSessionCache(r.Context(), currentSession.UserId, currentSession.Id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Response{
			"success" : true,
		})
	}
}


func EndAllOtherSessions(
	sessionStore postgres.SessionStoreInterface,
	cache redis.CacheInterface,
	) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claim :=  r.Context().Value(api.ContextClaimsJwt).(middleware.JWTClaims)
		userSessionList, err := sessionStore.GetSessionByUserId(claim.UserId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		for _, session := range userSessionList {
			if session.Id != claim.SessionId {
				err := sessionStore.DeleteSession(session)
				if err != nil {
					err = cache.DeleteSessionCache(r.Context(), claim.UserId, session.Id)
				}
			}
		}


		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Response{
			"success" : true,
		})
	}
}

func RefreshToken(
	jwt middleware.JwtHandlerInterface,
	) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claim :=  r.Context().Value(api.ContextClaimsJwt).(middleware.JWTClaims)
		refreshToken, err := jwt.GenerateRefreshToken(claim)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Response{
			"refresh_token" : refreshToken,
		})

	}
}

func NewAccessToken(jwt middleware.JwtHandlerInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claim :=  r.Context().Value(api.ContextClaimsJwt).(middleware.JWTClaims)
		if time.Unix(claim.ExpiresAt, 0).Sub(time.Now()) < api.AccessTokenExpTime {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		refreshToken, err := jwt.GenerateAccessToken(claim)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Response{
			"access_token" : refreshToken,
		})
	}
}