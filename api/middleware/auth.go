package middleware

import (
	"context"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
	"time"
	"userland/api"
	"userland/store/redis"
)



type AuthMiddleware struct {
	next http.Handler
	jwt  JwtHandlerInterface
	cache redis.CacheInterface
}



func NewAuthMiddleware(jwt JwtHandlerInterface, cache redis.CacheInterface) AuthMiddleware {
	return AuthMiddleware{
		jwt: jwt,
		cache: cache,
	}
}


func (authMiddleware AuthMiddleware) UserAuthMiddleware(next http.Handler) http.Handler {
	authMiddleware.next = next
	return authMiddleware
}

func (authMiddleware AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Key", "application/json")

	authHeader := r.Header.Get("Authorization")
	tokens := strings.Fields(authHeader)

	if len(tokens) < 2 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token := tokens[1]
	claim, err := authMiddleware.jwt.ClaimToken(token)

	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if claim == nil {
		log.Print(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if time.Unix(claim.ExpiresAt, 0).Sub(time.Now()) < 0 {
		log.Print(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if claim.UserHasTfa {
		if !claim.TfaVerified {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}

	session, err := authMiddleware.cache.GetSessionCache(r.Context(), claim.UserId, claim.SessionId)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if session.JwtId != claim.Id {
		log.Print(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ctx := context.WithValue(r.Context(), api.ContextClaimsJwt, *claim)


	authMiddleware.next.ServeHTTP(w, r.WithContext(ctx))
}