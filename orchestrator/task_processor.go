package orchestrator

import (
	"sync"
	"time"
)

type Task struct {
	ExpressionID  int
	TaskID        int           `json:"id"`
	Arg1          float64       `json:"arg1"`
	Arg2          float64       `json:"arg2"`
	Operation     string        `json:"operation"`
	OperationTime time.Duration `json:"operation_time"`
}

type Result struct {
	ExpressionID int
	TaskID       int     `json:"id"`
	Result       float64 `json:"result"`
}

type TaskManager struct {
	mu       sync.Mutex
	resultCh chan Result
	Counter  int
}

// NewTaskManager Создаёт новый экземпляр TaskManager
func NewTaskManager() *TaskManager {
	return &TaskManager{
		resultCh: make(chan Result, 1),
	}
}

// CreateTask Создаёт новую задачу
func (tm *TaskManager) CreateTask(arg1, arg2 float64, oper string, ExprID int) Task {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.Counter++
	taskID := tm.Counter
	operationTime := conf.GetOperationsTime(oper)
	task := Task{
		ExpressionID:  ExprID,
		TaskID:        taskID,
		Arg1:          arg1,
		Arg2:          arg2,
		Operation:     oper,
		OperationTime: operationTime,
	}
	return task
}

// AddResult Добавляет результат в канал
func (tm *TaskManager) AddResult(result Result) {
	tm.resultCh <- result
}

// GetResult Возвращает результат из канала
func (tm *TaskManager) GetResult() Result {
	return <-tm.resultCh
}
