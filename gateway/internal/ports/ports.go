package ports

import (
	"github.com/jaam8/web_calculator/common-lib/gen/auth_service"
	"github.com/jaam8/web_calculator/common-lib/gen/orchestrator"
)

type OrchestratorAdapter interface {
	Calculate(request *orchestrator.CalculateRequest) (*orchestrator.CalculateResponse, error)
	Expressions(request *orchestrator.ExpressionsRequest) (*orchestrator.ExpressionsResponse, error)
	ExpressionByID(request *orchestrator.ExpressionByIdRequest) (*orchestrator.ExpressionByIdResponse, error)
}

type AuthServiceAdapter interface {
	Login(request *auth_service.LoginRequest) (*auth_service.LoginResponse, error)
	Register(request *auth_service.RegisterRequest) (*auth_service.RegisterResponse, error)
	Refresh(request *auth_service.RefreshRequest) (*auth_service.RefreshResponse, error)
}
