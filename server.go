package main

import (
	"database/sql"
	"github.com/go-chi/chi/v5"
	"userland/api/auth"
	"userland/store/postgres"
	"userland/store/redis"
)

var router *chi.Mux

func InitServer(db *sql.DB) *chi.Mux {
	if router != nil{
		return router
	}


	redisClient := redis.InitRedisCache()
	cache := redis.NewRedisCacheStore(redisClient)
	authStore := postgres.NewAuthStore(cache)
	userStore := postgres.NewUserStore(db)
	router = chi.NewMux()

	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", auth.Register(userStore, authStore))
		r.Post("/verification", auth.RequestVerification(userStore, authStore))
		r.Get("/verify/{token}", auth.VerifyRegister(userStore, authStore))
	})

	return router
}