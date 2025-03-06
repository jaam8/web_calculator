package orchestrator

import (
	"sync"
)

type Expression struct {
	ID     int     `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

type ExpressionManager struct {
	mu          sync.Mutex
	expressions map[int]*Expression
	counter     int
}

// NewExpressionManager Создаёт новый экземпляр ExpressionManager
func NewExpressionManager() *ExpressionManager {
	return &ExpressionManager{
		expressions: make(map[int]*Expression),
	}
}

// CreateExpression Создаёт новую задачу и возвращает её ID
func (em *ExpressionManager) CreateExpression() (int, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.counter++
	expressionID := em.counter
	em.expressions[expressionID] = &Expression{
		ID:     expressionID,
		Status: "pending",
		//Queues: NewQueues(),
	}
	return expressionID, nil
}

// GetExpressions Возвращает все задачи
func (em *ExpressionManager) GetExpressions() []*Expression {
	var expressions []*Expression
	for _, expr := range em.expressions {
		expressions = append(expressions, expr)
	}
	return expressions
}

// GetExpression Возвращает задачу по ID
func (em *ExpressionManager) GetExpression(expressionID int) (*Expression, bool) {
	expr, exists := em.expressions[expressionID]
	return expr, exists
}

// ExpressionDone Завершает и обновляет статус задачи
func (em *ExpressionManager) ExpressionDone(expressionID int, result float64) {
	em.mu.Lock()
	defer em.mu.Unlock()
	expr, exists := em.expressions[expressionID]
	if exists {
		expr.Status = "done"
		expr.Result = result
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
		expr.Result = 0
		em.expressions[expressionID] = expr
	}
}
