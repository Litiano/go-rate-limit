package go_rate_limite

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Litiano/go-rate-limit/configs"
	ratelimiter "github.com/Litiano/go-rate-limit/infra/rate-limiter"
	"github.com/Litiano/go-rate-limit/infra/webserver/handlers"
	"github.com/Litiano/go-rate-limit/infra/webserver/middlewares"
	"github.com/redis/go-redis/v9"
)

func ClearDatabase(redisClient *redis.Client) {
	redisClient.FlushDB(context.Background())
}

func testRateLimit(t *testing.T, configureFn func(req *http.Request, config *configs.Config, rdb *redis.Client) int64) {
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "192.168.1.3:1234"

	config := configs.LoadConfig(".")
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.RedisHost, config.RedisPort),
		DB:       config.RedisDb + 1,
		Password: "",
	})
	ClearDatabase(rdb)

	limit := configureFn(req, config, rdb)

	redisRateLimiter := ratelimiter.NewRedisRateLimiter(rdb, config.DefaultRateLimit, time.Duration(config.BanTime)*time.Second)
	rateLimitMiddleware := middlewares.NewRateLimitMiddleware(redisRateLimiter, config.TokenAuthKey)
	handler := rateLimitMiddleware.RateLimitMiddleware(http.HandlerFunc(handlers.IndexHandler))

	for i := range limit * 2 {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if i < limit {
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}

			expected := `{"success":true}`
			result := strings.TrimSpace(rr.Body.String())
			if result != expected {
				t.Errorf("handler returned unexpected body: got %v want %v",
					rr.Body.String(), expected)
			}
		} else {
			if status := rr.Code; status != http.StatusTooManyRequests {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusTooManyRequests)
			}

			expected := `you have reached the maximum number of requests or actions allowed within a certain time frame`
			result := strings.TrimSpace(rr.Body.String())
			if result != expected {
				t.Errorf("handler returned unexpected body: got %v want %v",
					rr.Body.String(), expected)
			}
		}
	}
}

func TestDefaultRateLimitWithUser(t *testing.T) {
	testRateLimit(t, func(req *http.Request, config *configs.Config, rdb *redis.Client) int64 {
		_, tokenString, err := config.TokenAuthKey.Encode(map[string]interface{}{
			"sub": "commonUserTest",
			"exp": time.Now().Add(time.Second * time.Duration(config.JwtExpiresIn)).Unix(),
		})
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("API_KEY", tokenString)

		return config.DefaultRateLimit
	})
}

func TestDefaultRateLimitWithoutUser(t *testing.T) {
	testRateLimit(t, func(req *http.Request, config *configs.Config, rdb *redis.Client) int64 {
		return config.DefaultRateLimit
	})
}

func TestSuperRateLimit(t *testing.T) {
	limit := int64(30)
	testRateLimit(t, func(req *http.Request, config *configs.Config, rdb *redis.Client) int64 {
		_, tokenString, err := config.TokenAuthKey.Encode(map[string]interface{}{
			"sub": "superUserTest",
			"exp": time.Now().Add(time.Second * time.Duration(config.JwtExpiresIn)).Unix(),
		})
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("API_KEY", tokenString)
		rdb.Set(context.Background(), "rate-limit-for:superUserTest", limit, 0)

		return limit
	})
}
