package configs

import (
	"github.com/go-chi/jwtauth"
	"github.com/spf13/viper"
)

type Config struct {
	RedisHost        string `mapstructure:"REDIS_HOST"`
	RedisPort        int    `mapstructure:"REDIS_PORT"`
	RedisDb          int    `mapstructure:"REDIS_DB"`
	AppPort          int    `mapstructure:"APP_PORT"`
	JwtSecret        string `mapstructure:"JWT_SECRET"`
	JwtExpiresIn     int    `mapstructure:"JWT_EXPIRES_IN"`
	DefaultRateLimit int64  `mapstructure:"DEFAULT_RATE_LIMIT"`
	BanTime          int    `mapstructure:"BAN_TIME"`
	TokenAuthKey     *jwtauth.JWTAuth
}

var cfg *Config

func LoadConfig(path string) *Config {
	viper.SetConfigName("app_config")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		panic(err)
	}
	cfg.TokenAuthKey = jwtauth.New("HS256", []byte(cfg.JwtSecret), nil)

	return cfg
}
