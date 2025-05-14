package ports

import (
	"github.com/google/uuid"
	"github.com/jaam8/web_calculator/orchestrator/internal/models"
)

type StorageAdapter interface {
	SaveExpression(expression models.Expression) (uuid.UUID, error)
	GetExpressionById(userId uuid.UUID, id uuid.UUID) (*models.Expression, error)
	GetExpressions(userId uuid.UUID) ([]*models.Expression, error)
	UpdateExpression(userId uuid.UUID, id uuid.UUID, status *string, result *float64) error
}
