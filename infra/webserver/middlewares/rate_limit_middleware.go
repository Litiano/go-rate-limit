package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	ratelimiter "github.com/Litiano/go-rate-limit/infra/rate-limiter"
	"github.com/go-chi/jwtauth"
)

type RateLimitMiddleware struct {
	limiter ratelimiter.RateLimiterInterface
}

func NewRateLimitMiddleware(limiter ratelimiter.RateLimiterInterface) *RateLimitMiddleware {
	return &RateLimitMiddleware{limiter: limiter}
}

func (rl *RateLimitMiddleware) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		jwtAuth := request.Context().Value("jwt").(*jwtauth.JWTAuth)
		subject := strings.Split(request.RemoteAddr, ":")[0]

		token := request.Header.Get("API_KEY")
		if token != "" {
			data, err := jwtAuth.Decode(token)
			if err != nil {
				response.WriteHeader(http.StatusUnauthorized)
				response.Write([]byte(err.Error()))
				return
			}
			subject = data.Subject()
		}
		fmt.Printf("RateLimitMiddleware for %s\n", subject)

		banned, err := rl.limiter.IsBanned(subject)
		if err != nil {
			return400(response, err)
			return
		}
		if banned {
			return429(response)
			return
		}

		rateLimit, err := rl.limiter.GetRateLimitFor(subject)
		if err != nil {
			return400(response, err)
			return
		}

		count, err := rl.limiter.GetRateCountFor(subject)
		if err != nil {
			return400(response, err)
			return
		}

		if count > rateLimit {
			err = rl.limiter.Ban(subject)
			if err != nil {
				return400(response, err)
			}
			return429(response)
			return
		}

		next.ServeHTTP(response, request)
	})
}

func return400(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(err.Error()))
}

func return429(w http.ResponseWriter) {
	w.WriteHeader(http.StatusTooManyRequests)
	w.Write([]byte("you have reached the maximum number of requests or actions allowed within a certain time frame"))
}
