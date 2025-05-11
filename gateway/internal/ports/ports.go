package ports

import "github.com/jaam8/web_calculator/common-lib/gen/orchestrator"

type OrchestratorAdapter interface {
	Calculate(request *orchestrator.CalculateRequest) (*orchestrator.CalculateResponse, error)
	Expressions(request *orchestrator.ExpressionsRequest) (*orchestrator.ExpressionsResponse, error)
	ExpressionByID(request *orchestrator.ExpressionByIdRequest) (*orchestrator.ExpressionByIdResponse, error)
}
