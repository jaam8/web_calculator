package types

import (
	"github.com/google/uuid"
	"github.com/jaam8/web_calculator/orchestrator/internal/models"
)

type TaskManager interface {
	CreateTask(arg1 float64, arg2 float64, oper string, ExprID uuid.UUID) models.Task
	AddResult(result models.Result)
	GetResult() models.Result
}

type ExpressionManager interface {
	CreateExpression(expression *models.Expression) error
	GetTaskManager(expressionID uuid.UUID) (TaskManager, error)
	GetExpressions() []*models.Expression
	GetExpression(expressionID uuid.UUID) (*models.Expression, bool)
	AddTask(task models.Task)
	GetTasks() chan models.Task
	ExpressionDone(expressionID uuid.UUID, result float64)
	ExpressionError(expressionID uuid.UUID)
}
