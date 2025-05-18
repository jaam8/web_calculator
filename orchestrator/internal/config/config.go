package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jaam8/web_calculator/common-lib/postgres"
	"time"
)

type OrchestratorConfig struct {
	Host string `yaml:"host" env:"HOST" env-default:"localhost"`
	Port int    `yaml:"port" env:"PORT" env-default:"50052"`

	TimeAddition        int `env:"TIME_ADDITION_MS"`
	TimeSubtraction     int `env:"TIME_SUBTRACTION_MS"`
	TimeMultiplications int `env:"TIME_MULTIPLICATIONS_MS"`
	TimeDivisions       int `env:"TIME_DIVISIONS_MS"`
}

type Config struct {
	Orchestrator  OrchestratorConfig `yaml:"orchestrator" env-prefix:"ORCHESTRATOR_"`
	Postgres      postgres.Config    `yaml:"postgres" env-prefix:"POSTGRES_"`
	LogLevel      string             `yaml:"log_level" env:"LOG_LEVEL" env-default:"info"`
	MigrationPath string             `yaml:"migration_path" env:"MIGRATION_PATH" env-default:"file:///db/migrations"`
}

func New() (Config, error) {
	var cfg Config
	// docker workdir - app/
	// local workdir - web_calculator/orchestrator
	if err := cleanenv.ReadConfig("../.env", &cfg); err != nil {
		fmt.Println(err.Error())
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return Config{}, fmt.Errorf("failed to read env vars: %v", err)
		}
	}

	return cfg, nil
}

func (c OrchestratorConfig) GetOperationsTime(oper string) time.Duration {
	switch oper {
	case "+":
		return time.Duration(c.TimeAddition) * time.Millisecond
	case "-":
		return time.Duration(c.TimeSubtraction) * time.Millisecond
	case "*":
		return time.Duration(c.TimeMultiplications) * time.Millisecond
	case "/":
		return time.Duration(c.TimeDivisions) * time.Millisecond
	default:
		return 0
	}
}
