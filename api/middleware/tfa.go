package middleware

import (
	"net/http"
	"userland/api"
)

type TfaRequired struct {
	Next http.Handler
}

func (middleware TfaRequired) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	claim := r.Context().Value(api.ContextClaimsJwt).(JWTClaims)

	if claim.UserHasTfa {
		if !claim.TfaVerified {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

	}

	middleware.Next.ServeHTTP(w, r)
}

func TfaRequiredMiddleware(next http.Handler) http.Handler {
	return TfaRequired{Next: next}
}
