package middleware

import "net/http"

type RequestMustBeJson struct {
	Next http.Handler
}

func (middleware RequestMustBeJson) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	middleware.Next.ServeHTTP(w, r)
}
