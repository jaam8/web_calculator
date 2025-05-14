package ports

import (
	"github.com/jaam8/web_calculator/agent/internal/models"
)

type OrchestratorAdapter interface {
	GetTask() (models.Task, error)
	ResultTask(expressionID string, taskID int, result float64) (string, error)
}
