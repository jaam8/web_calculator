module github.com/jaam8/web_calculator/agent

go 1.24.2

require (
	github.com/ilyakaznacheev/cleanenv v1.5.0
	github.com/jaam8/web_calculator/common-lib v0.0.0-20250508175247-20d2b45e5adc
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.72.0
	google.golang.org/protobuf v1.36.6
)

replace github.com/jaam8/web_calculator/common-lib => ../common-lib

require (
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250505200425-f936aa4a68b2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)
