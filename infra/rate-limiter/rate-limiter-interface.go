package rate_limiter

type RateLimiterInterface interface {
	GetRateLimitFor(subject string) (int64, error)
	GetRateCountFor(subject string) (int64, error)
	IsBanned(subject string) (bool, error)
	Ban(subject string) error
	Unban(subject string) error
}
