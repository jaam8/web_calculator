package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jaam8/web_calculator/common-lib/postgres"
	"github.com/jaam8/web_calculator/common-lib/redis"
)

type AuthServiceConfig struct {
	Host string `yaml:"host" env:"HOST" env-default:"localhost"`
	Port int    `yaml:"port" env:"PORT" env-default:"50051"`

	RedisDB      int    `yaml:"redis_db" env:"REDIS_DB" env-default:"0"`
	UpstreamName string `yaml:"upstream_name" env:"UPSTREAM_NAME" env-default:"auth_service"`
	UpstreamPort int    `yaml:"upstream_port" env:"UPSTREAM_PORT" env-default:"50051"`
}

type Config struct {
	AuthService AuthServiceConfig `yaml:"auth_service" env-prefix:"AUTH_SERVICE_"`
	Redis       redis.Config      `yaml:"redis" env-prefix:"REDIS_"`
	Postgres    postgres.Config   `yaml:"postgres" env-prefix:"POSTGRES_"`

	LogLevel          string `yaml:"log_level" env:"LOG_LEVEL" env-default:"info"`
	JwtSecret         string `yaml:"jwt_secret" env:"JWT_SECRET"`
	RefreshExpiration int    `yaml:"refresh_expiration" env:"REFRESH_EXPIRATION" env-default:"24"`
	AccessExpiration  int    `yaml:"access_expiration" env:"ACCESS_EXPIRATION" env-default:"15"`
	MigrationPath     string `yaml:"migration_path" env:"MIGRATION_PATH" env-default:"file:///db/migrations"`
}

func New() (Config, error) {
	var cfg Config
	// docker workdir - app/
	// local workdir - web_calculator/auth_service
	if err := cleanenv.ReadConfig("../.env", &cfg); err != nil {
		fmt.Println(err.Error())
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return Config{}, fmt.Errorf("failed to read env vars: %v", err)
		}
	}

	return cfg, nil
}
