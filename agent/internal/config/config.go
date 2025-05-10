package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type OrchestratorConfig struct {
	Host string `yaml:"host" env:"HOST" env-default:"localhost"`
	Port int    `yaml:"port" env:"PORT" env-default:"50052"`

	Timeout        int  `yaml:"timeout" env:"TIMEOUT_MS" env-default:"500"`
	MaxRetries     uint `yaml:"max_retries" env:"MAX_RETRIES" env-default:"3"`
	BaseRetryDelay int  `yaml:"base_retry_delay" env:"BASE_RETRY_DELAY" env-default:"100"`
}

type AgentConfig struct {
	Host string `yaml:"host" env:"HOST" env-default:"localhost"`
	Port int    `yaml:"port" env:"PORT" env-default:"50051"`

	ComputingPower int `yaml:"computing_power" env:"COMPUTING_POWER" env-default:"5"`
	WaitTime       int `yaml:"wait_time" env:"WAIT_TIME_MS" env-default:"500"`
}

type Config struct {
	Orchestrator OrchestratorConfig `yaml:"orchestrator" env-prefix:"ORCHESTRATOR"`
	Agent        AgentConfig        `yaml:"agent" env-prefix:"AGENT"`
}

func New() (Config, error) {
	var cfg Config
	// docker workdir - app/
	// local workdir - web_calculator/agent
	if err := cleanenv.ReadConfig("../.env", &cfg); err != nil {
		fmt.Println(err.Error())
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return Config{}, fmt.Errorf("failed to read env vars: %v", err)
		}
	}

	return cfg, nil
}
