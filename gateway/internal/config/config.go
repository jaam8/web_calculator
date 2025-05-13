package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type OrchestratorConfig struct {
	Host string `yaml:"host" env:"HOST" env-default:"localhost"`
	Port int    `yaml:"port" env:"PORT" env-default:"50052"`

	UpstreamName string `yaml:"upstream_name" env:"UPSTREAM_NAME" env-default:"orchestrator"`
	UpstreamPort int    `yaml:"upstream_port" env:"UPSTREAM_PORT" env-default:"50052"`

	Timeout        int  `yaml:"timeout" env:"TIMEOUT_MS" env-default:"500"`
	MaxRetries     uint `yaml:"max_retries" env:"MAX_RETRIES" env-default:"3"`
	BaseRetryDelay int  `yaml:"base_retry_delay" env:"BASE_RETRY_DELAY" env-default:"100"`
}

type AuthConfig struct {
	Host           string `yaml:"host" env:"HOST" env-default:"localhost"`
	Port           int    `yaml:"port" env:"PORT" env-default:"50051"`
	UpstreamName   string `yaml:"upstream_name" env:"UPSTREAM_NAME" env-default:"auth"`
	UpstreamPort   int    `yaml:"upstream_port" env:"UPSTREAM_PORT" env-default:"50051"`
	Timeout        int    `yaml:"timeout" env:"TIMEOUT_MS" env-default:"500"`
	MaxRetries     uint   `yaml:"max_retries" env:"MAX_RETRIES" env-default:"3"`
	BaseRetryDelay int    `yaml:"base_retry_delay" env:"BASE_RETRY_DELAY" env-default:"100"`
}

type GrpcPoolConfig struct {
	MaxConns         int  `yaml:"max_conns" env:"MAX_CONNS" env-default:"10"`
	MinConns         int  `yaml:"min_conns" env:"MIN_CONNS" env-default:"1"`
	MaxRetries       uint `yaml:"max_retries" env:"MAX_RETRIES" env-default:"3"`
	BaseRetryDelayMs uint `yaml:"base_retry_delay_ms" env:"BASE_RETRY_DELAY_MS" env-default:"200"`
}

type GatewayConfig struct {
	Host string `yaml:"host" env:"HOST" env-default:"localhost"`
	Port int    `yaml:"port" env:"PORT" env-default:"8080"`
}

type Config struct {
	Orchestrator OrchestratorConfig `yaml:"orchestrator" env-prefix:"ORCHESTRATOR_"`
	Gateway      GatewayConfig      `yaml:"gateway" env-prefix:"GATEWAY_"`
	GrpcPool     GrpcPoolConfig     `yaml:"grpc_pool" env-prefix:"GRPC_POOL_"`
	AuthService  AuthConfig         `yaml:"auth_service" env-prefix:"AUTH_SERVICE_"`
	LogLevel     string             `yaml:"log_level" env:"LOG_LEVEL" env-default:"info"`
	JwtSecret    string             `yaml:"jwt_secret" env:"JWT_SECRET"`
	AccessTTL    int                `yaml:"access_ttl" env:"ACCESS_EXPIRATION" env-default:"15"`
	RefreshTTL   int                `yaml:"refresh_ttl" env:"REFRESH_EXPIRATION" env-default:"24"`
}

func New() (Config, error) {
	var cfg Config
	// docker workdir - app/
	// local workdir - web_calculator/gateway
	if err := cleanenv.ReadConfig("../.env", &cfg); err != nil {
		fmt.Println(err.Error())
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return Config{}, fmt.Errorf("failed to read env vars: %v", err)
		}
	}

	return cfg, nil
}
