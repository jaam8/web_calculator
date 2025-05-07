package orchestrator

import (
	"sync"
)

type Expression struct {
	ID     int      `json:"id"`
	Status string   `json:"status"`
	Result *float64 `json:"result,omitempty"`
}

type ExpressionManager struct {
	mu           sync.Mutex
	expressions  map[int]*Expression
	taskManagers map[int]*TaskManager
	TaskCh       chan Task
	counter      int
}

// NewExpressionManager Создаёт новый экземпляр ExpressionManager
func NewExpressionManager() *ExpressionManager {
	return &ExpressionManager{
		expressions:  make(map[int]*Expression),
		taskManagers: make(map[int]*TaskManager),
		TaskCh:       make(chan Task, 100),
	}
}

// CreateExpression Создаёт новую задачу и возвращает её TaskID
func (em *ExpressionManager) CreateExpression() (int, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.counter++
	expressionID := em.counter
	em.expressions[expressionID] = &Expression{
		ID:     expressionID,
		Status: "pending",
	}
	em.taskManagers[expressionID] = NewTaskManager()
	return expressionID, nil
}

// GetTaskManager Возвращает TaskManager по ExpressionID
func (em *ExpressionManager) GetTaskManager(expressionID int) (*TaskManager, bool) {
	taskManager, exists := em.taskManagers[expressionID]
	if !exists {
		return nil, false
	}
	return taskManager, true
}

// GetExpressions Возвращает все таски
func (em *ExpressionManager) GetExpressions() []*Expression {
	var expressions []*Expression
	for _, expr := range em.expressions {
		expressions = append(expressions, expr)
	}
	return expressions
}

// GetExpression Возвращает задачу по TaskID
func (em *ExpressionManager) GetExpression(expressionID int) (*Expression, bool) {
	expr, exists := em.expressions[expressionID]
	return expr, exists
}

// AddTask Добавляет задачу в очередь на вычисление
func (em *ExpressionManager) AddTask(task Task) {
	em.TaskCh <- task
}

// GetTasks Возвращает канал с задачами
func (em *ExpressionManager) GetTasks() chan Task {
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
