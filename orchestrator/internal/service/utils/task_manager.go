package utils

import (
	"github.com/google/uuid"
	"github.com/jaam8/web_calculator/orchestrator/internal/models"
	"sync"
	"time"
)

type TaskManager struct {
	durations map[string]int
	mu        sync.Mutex
	resultCh  chan models.Result
	Counter   int
}

// NewTaskManager Создаёт новый экземпляр TaskManager
func NewTaskManager(durations map[string]int) *TaskManager {
	return &TaskManager{
		durations: durations,
		resultCh:  make(chan models.Result, 1),
	}
}

// CreateTask Создаёт новую задачу
func (tm *TaskManager) CreateTask(arg1, arg2 float64, oper string, ExprID uuid.UUID) models.Task {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.Counter++
	taskID := tm.Counter
	operationTime := tm.durations[oper]
	task := models.Task{
		ExpressionID:  ExprID,
		TaskID:        taskID,
		Arg1:          arg1,
		Arg2:          arg2,
		Operation:     oper,
		OperationTime: time.Millisecond * time.Duration(operationTime),
	}
	return task
}

// AddResult Добавляет результат в канал
func (tm *TaskManager) AddResult(result models.Result) {
	tm.resultCh <- result
}

// GetResult Возвращает результат из канала
func (tm *TaskManager) GetResult() models.Result {
	return <-tm.resultCh
}
