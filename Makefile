generate_orchestrator:
	protoc -I ./orchestrator \
		--go_out=. \
		--go-grpc_out=. \
		./orchestrator/api/orchestrator.proto

generate_auth_service:
	protoc -I ./auth_service \
		--go_out=. \
		--go-grpc_out=. \
		./auth_service/api/auth_service.proto

.PHONY: run

run:
	@echo "⚙️  Running all services locally..."
	@echo "💡 Make sure you’ve run: make proto && migrations && exported configs before this"
	@$(MAKE) -j 5 start-auth start-orchestrator start-agent start-gateway start-frontend

start-auth:
	@echo "🚀 Starting auth_service..."
	go run ./auth_service/cmd/main.go

start-orchestrator:
	@echo "🚀 Starting orchestrator..."
	go run ./orchestrator/cmd/main.go

start-agent:
	@echo "🚀 Starting agent..."
	go run ./agent/cmd/main.go

start-gateway:
	@echo "🚀 Starting gateway..."
	go run ./gateway/cmd/main.go

start-frontend:
	@echo "🚀 Starting frontend..."
	go run ./frontend/main.go
