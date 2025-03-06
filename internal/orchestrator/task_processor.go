package orchestrator

import (
	"sync"
	"time"
)

type Task struct {
	ID            int           `json:"id"`
	Arg1          float64       `json:"arg1"`
	Arg2          float64       `json:"arg2"`
	Operation     string        `json:"operation"`
	OperationTime time.Duration `json:"operation_time"`
}

type Result struct {
	ID     int     `json:"id"`
	Result float64 `json:"result"`
}

type TaskManager struct {
	mu          sync.Mutex
	tasks       map[int]Task
	taskQueues  chan Task
	resultQueue chan Result
	Counter     int
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks:       make(map[int]Task),
		taskQueues:  make(chan Task, 10),
		resultQueue: make(chan Result, 10),
	}
}

func (tm *TaskManager) CreateTask(arg1, arg2 float64, oper string) Task {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.Counter++
	taskID := tm.Counter
	operationTime := conf.GetOperationsTime(oper)
	task := Task{
		ID:            taskID,
		Arg1:          arg1,
		Arg2:          arg2,
		Operation:     oper,
		OperationTime: operationTime,
	}
	tm.tasks[taskID] = task
	return task
}

func (tm *TaskManager) AddTask(task Task) {
	tm.taskQueues <- task

}

func (tm *TaskManager) GetTask(taskID int) (Task, bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	task, exists := tm.tasks[taskID]
	return task, exists

}

func (tm *TaskManager) GetTasksChan() chan Task {
	return tm.taskQueues
}

func (tm *TaskManager) AddResult(result Result) {
	tm.resultQueue <- result
}

func (tm *TaskManager) GetResult() Result {
	return <-tm.resultQueue
}

func (tm *TaskManager) RemoveTask(taskID int) bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	delete(tm.tasks, taskID)
	_, exists := tm.tasks[taskID]
	return exists
}
