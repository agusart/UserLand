package main

import (
	"github.com/go-chi/chi/v5"
	"userland/api/auth"
	"userland/store/postgres"
	"userland/store/redis"
)

var router *chi.Mux

func InitServer() *chi.Mux {
	if router != nil{
		return router
	}

	db := postgres.NewDatabase()
	redisClient := redis.InitRedisDb()
	cache := redis.NewRedisCacheStore(redisClient)

	userStore := postgres.NewUserStore(db, cache)

	router = chi.NewMux()

	router.Route("/auth", func(r chi.Router) {
		router.Post("/register", auth.Register(userStore))
		router.Post("/verification", auth.Verification(userStore))
	})

	return router
}