package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Litiano/go-rate-limit/configs"
	ratelimiter "github.com/Litiano/go-rate-limit/infra/rate-limiter"
	"github.com/Litiano/go-rate-limit/infra/webserver/handlers"
	"github.com/Litiano/go-rate-limit/infra/webserver/middlewares"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)
import "github.com/redis/go-redis/v9"

func main() {
	config := configs.LoadConfig(".")
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.RedisHost, config.RedisPort),
		DB:       config.RedisDb,
		Password: "",
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected to Redis!")

	// Set superUser rate limit
	rdb.Set(ctx, "rate-limit-for:superUser", 100, 0)
	//

	authHandler := handlers.NewAuthHandler(config.TokenAuthKey, config.JwtExpiresIn)
	redisRateLimite := ratelimiter.NewRedisRateLimiter(rdb, config.DefaultRateLimit, time.Duration(config.BanTime)*time.Second)
	rateLimiter := middlewares.NewRateLimitMiddleware(redisRateLimite)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RealIP)
	router.Use(middleware.WithValue("jwt", config.TokenAuthKey))
	router.Use(rateLimiter.RateLimitMiddleware)

	router.Group(func(router chi.Router) {
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
			})
		})
	})

	// Open routes
	router.Group(func(router chi.Router) {
		router.Post("/auth/login", authHandler.LoginHandler)
	})

	err = http.ListenAndServe(fmt.Sprintf(":%d", config.AppPort), router)
	if err != nil {
		return
	}
}
