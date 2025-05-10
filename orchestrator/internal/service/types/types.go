package types

import (
	"github.com/jaam8/web_calculator/orchestrator/internal/models"
)

type TaskManager interface {
	CreateTask(arg1 float64, arg2 float64, oper string, ExprID int) models.Task
	AddResult(result models.Result)
	GetResult() models.Result
}

type ExpressionManager interface {
	CreateExpression() (int, error)
	GetTaskManager(expressionID int) (TaskManager, bool)
	GetExpressions() []*models.Expression
	GetExpression(expressionID int) (*models.Expression, bool)
	AddTask(task models.Task)
	GetTasks() chan models.Task
	ExpressionDone(expressionID int, result float64)
	ExpressionError(expressionID int)
}
