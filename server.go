package main

import (
	"database/sql"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
	"userland/api"
	"userland/api/auth"
	"userland/api/middleware"
	"userland/api/session"
	"userland/store/postgres"
	"userland/store/redis"
)

var router *chi.Mux

func InitServer(db *sql.DB) *chi.Mux {
	if router != nil{
		return router
	}

	jwtConfig := middleware.NewJWTConfig(
			"asdf",
			api.RefreshTokenExpTime,
			api.AccessTokenExpTime,
			jwt.SigningMethodHS256,
		)

	jwtHandler := middleware.NewJWTHandler(jwtConfig)

	redisClient := redis.InitRedisCache()
	cache := redis.NewRedisCacheStore(redisClient)

	authStore := postgres.NewAuthStore(cache)
	userStore := postgres.NewUserStore(db)
	sessionStore := postgres.NewSessionStore(db)

	router = chi.NewMux()
	authMiddleware := middleware.NewAuthMiddleware(jwtHandler, cache)

	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", auth.Register(userStore, authStore))
		r.Post("/verification", auth.RequestVerification(userStore, authStore))
		r.Get("/verify/{token}", auth.VerifyRegister(userStore, authStore))
		r.Post("/password/forgot", auth.ForgetPassword(userStore, authStore))
		r.Post("/password/reset", auth.ResetPassword(userStore, authStore))
		r.Post("/login", auth.Login(userStore, jwtHandler, sessionStore, authStore, cache))

	})

	router.Route("/me", func(r chi.Router) {
		r.Use(authMiddleware.UserAuthMiddleware)
		r.Get("/session", session.ListSession(sessionStore))
	})

	return router
}