package utils

import (
	"github.com/jaam8/web_calculator/orchestrator/internal/models"
	"github.com/jaam8/web_calculator/orchestrator/internal/service/types"
	"sync"
)

type ExpressionManager struct {
	mu           sync.Mutex
	expressions  map[int]*models.Expression
	taskManagers map[int]*TaskManager
	TaskCh       chan models.Task
	counter      int
	durations    map[string]int
}

// NewExpressionManager Создаёт новый экземпляр ExpressionManager
func NewExpressionManager(durations map[string]int) *ExpressionManager {
	return &ExpressionManager{
		expressions:  make(map[int]*models.Expression),
		taskManagers: make(map[int]*TaskManager),
		TaskCh:       make(chan models.Task, 100),
		durations:    durations,
	}
}

// CreateExpression Создаёт новую задачу и возвращает её TaskID
func (em *ExpressionManager) CreateExpression() (int, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.counter++
	expressionID := em.counter
	em.expressions[expressionID] = &models.Expression{
		ExpressionID: expressionID,
		Status:       "pending",
	}
	em.taskManagers[expressionID] = NewTaskManager(em.durations)
	return expressionID, nil
}

// GetTaskManager Возвращает TaskManager по ExpressionID
func (em *ExpressionManager) GetTaskManager(expressionID int) (types.TaskManager, bool) {
	taskManager, exists := em.taskManagers[expressionID]
	if !exists || taskManager == nil {
		return taskManager, false
	}
	return taskManager, true
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
func (em *ExpressionManager) GetExpression(expressionID int) (*models.Expression, bool) {
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
func (em *ExpressionManager) ExpressionDone(expressionID int, result float64) {
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
func (em *ExpressionManager) ExpressionError(expressionID int) {
	em.mu.Lock()
	defer em.mu.Unlock()
	expr, exists := em.expressions[expressionID]
	if exists {
		expr.Status = "invalid expression"
		em.expressions[expressionID] = expr
	}
}
