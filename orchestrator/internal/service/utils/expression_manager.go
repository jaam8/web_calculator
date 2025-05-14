package utils

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/jaam8/web_calculator/orchestrator/internal/models"
	"github.com/jaam8/web_calculator/orchestrator/internal/service/types"
	"sync"
)

type ExpressionManager struct {
	mu           sync.Mutex
	expressions  map[uuid.UUID]*models.Expression
	taskManagers map[uuid.UUID]*TaskManager
	TaskCh       chan models.Task
	counter      int
	durations    map[string]int
}

// NewExpressionManager Создаёт новый экземпляр ExpressionManager
func NewExpressionManager(durations map[string]int) *ExpressionManager {
	return &ExpressionManager{
		expressions:  make(map[uuid.UUID]*models.Expression),
		taskManagers: make(map[uuid.UUID]*TaskManager),
		TaskCh:       make(chan models.Task, 100),
		durations:    durations,
	}
}

// CreateExpression Создаёт TaskManager и добавляет его в мапу, а также добавляет Expression в мапу
func (em *ExpressionManager) CreateExpression(expression *models.Expression) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.expressions[expression.ExpressionID] = expression
	em.taskManagers[expression.ExpressionID] = NewTaskManager(em.durations)
	return nil
}

// GetTaskManager Возвращает TaskManager по ExpressionID
func (em *ExpressionManager) GetTaskManager(expressionID uuid.UUID) (types.TaskManager, error) {
	taskManager, exists := em.taskManagers[expressionID]
	if !exists || taskManager == nil {
		return taskManager, fmt.Errorf("task manager not found for expression ID: %s", expressionID)
	}
	return taskManager, nil
}

// GetExpressions Возвращает все таски
func (em *ExpressionManager) GetExpressions() []*models.Expression {
	var expressions []*models.Expression
	for _, expr := range em.expressions {
		expressions = append(expressions, expr)
	}
	return expressions
}

// GetExpression Возвращает задачу по TaskID
func (em *ExpressionManager) GetExpression(expressionID uuid.UUID) (*models.Expression, bool) {
	expr, exists := em.expressions[expressionID]
	return expr, exists
}

// AddTask Добавляет задачу в очередь на вычисление
func (em *ExpressionManager) AddTask(task models.Task) {
	em.TaskCh <- task
}

// GetTasks Возвращает канал с задачами
func (em *ExpressionManager) GetTasks() chan models.Task {
	return em.TaskCh
}

// ExpressionDone Завершает и обновляет статус задачи
func (em *ExpressionManager) ExpressionDone(expressionID uuid.UUID, result float64) {
	em.mu.Lock()
	defer em.mu.Unlock()
	expr, exists := em.expressions[expressionID]
	if exists {
		expr.Status = "done"
		expr.Result = &result
		em.expressions[expressionID] = expr
	}
}

// ExpressionError ставит ошибку в статусе если вдруг задача прошадшая валидацию, оказалась с ошибкой
func (em *ExpressionManager) ExpressionError(expressionID uuid.UUID) {
	em.mu.Lock()
	defer em.mu.Unlock()
	expr, exists := em.expressions[expressionID]
	if exists {
		expr.Status = "invalid expression"
		em.expressions[expressionID] = expr
	}
}
