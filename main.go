package main

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"net/http"
	"os"
	"strconv"
	"userland/api"
	"userland/api/me"
	"userland/api/middleware"
	"userland/api/worker"
	"userland/store/broker"
	"userland/store/postgres"
	"userland/store/redis"
)

func main() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	err := godotenv.Load()
	if err != nil {
		log.Err(err)
	}

	port, err := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	if err != nil {
		log.Err(err)
	}
	postgresCfg := postgres.PGConfig{
		Host:     os.Getenv("POSTGRES_ADDR"),
		Port:     port,
		Username: os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Database: os.Getenv("POSTGRES_DB"),
	}

	db, err := postgres.NewPG(postgresCfg)
	if err != nil {
		log.Err(err)
	}

	jwtConfig := middleware.NewJWTConfig(
		os.Getenv("JWT_KEY"),
		api.RefreshTokenExpTime,
		api.AccessTokenExpTime,
		jwt.SigningMethodHS256,
	)

	jwtHandler := middleware.NewJWTHandler(jwtConfig)
	redisClient := redis.InitRedisCache()
	cache := redis.NewRedisCacheStore(redisClient)
	authStore := redis.NewAuthStore(cache)
	userStore := postgres.NewUserStore(db)
	sessionStore := postgres.NewSessionStore(db)
	tfaStore := postgres.NewTfaStore(db)
	logStore := postgres.NewLogStore(db)
	fileHelper := me.FileHelper{}

	consumerBrokerConfig := &kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("BOOTSTRAPSERVERS"),
		"group.id":          "userland",
	}

	producerBrokerConfig := &kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("BOOTSTRAPSERVERS"),
	}

	msgBroker, err := broker.NewBroker(consumerBrokerConfig, producerBrokerConfig)
	if err !=nil {
		panic(err)
	}

	r := InitRouter(
		jwtHandler,
		cache,
		fileHelper,
		userStore,
		authStore,
		sessionStore,
		tfaStore,
		msgBroker,
		)

	endWorkerChan := make(chan int, 1)
	defer func() {
		endWorkerChan <- 1
	}()

	go worker.UserLoginLog(context.Background(), msgBroker, logStore, endWorkerChan)

	_ = http.ListenAndServe(":8080", r)
}
