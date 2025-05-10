package service

import (
	"errors"
	"github.com/jaam8/web_calculator/agent/internal/models"
	"github.com/jaam8/web_calculator/agent/internal/ports"
	errs "github.com/jaam8/web_calculator/common-lib/errors"
	"log"
	"time"
)

type AgentService struct {
	orchestratorAdapter ports.OrchestratorAdapter
}

func NewAgentService(orchestratorAdapter ports.OrchestratorAdapter) *AgentService {
	return &AgentService{
		orchestratorAdapter: orchestratorAdapter,
	}
}

// Work делает постоянные запросы к оркестратору
func (s *AgentService) Work(waitTime int) {
	sleepTime := time.Duration(waitTime) * time.Millisecond
	for {
		task, err := s.GetTask()
		for err != nil {
			if errors.Is(err, errs.ErrTaskNotFound) {
				log.Println("Task not found:", err)
				time.Sleep(sleepTime)
			}
			if errors.Is(err, errs.ErrInternalServerError) {
				log.Println("Internal server error:", err)
				time.Sleep(sleepTime)
			}
			task, err = s.GetTask()
		}
		result, err := DoTask(task)
		if err != nil {
			log.Println("Error calculating task:", err)
			if errors.Is(err, errs.ErrDivideByZero) {
				log.Println("Division by zero error:", err)
				continue
			}
			if errors.Is(err, errs.ErrInvalidExpression) {
				log.Println("Invalid expression error:", err)
				continue
			}

		}

		Result := models.Result{
			ExpressionID: task.ExpressionID,
			TaskID:       task.TaskID,
			Result:       result,
		}

		err = s.PostResult(Result)
		if err != nil {
			log.Println("Error posting result:", err)
		}
		log.Printf("Task TaskID: %d processed, result: %f\n", task.TaskID, result)
		time.Sleep(sleepTime)
	}
}

// DoTask вычисляет задачу
func DoTask(task models.Task) (float64, error) {
	switch task.Operation {
	case "+":
		time.Sleep(task.OperationTime)
		return task.Arg1 + task.Arg2, nil
	case "-":
		time.Sleep(task.OperationTime)
		return task.Arg1 - task.Arg2, nil
	case "*":
		time.Sleep(task.OperationTime)
		return task.Arg1 * task.Arg2, nil
	case "/":
		time.Sleep(task.OperationTime)
		if task.Arg2 == 0 {
			return 0, errs.ErrDivideByZero
		}
		return task.Arg1 / task.Arg2, nil
	default:
		return 0, errs.ErrInvalidExpression
	}
}

// GetTask делает запрос к оркестратору и возвращает задачу
func (s *AgentService) GetTask() (models.Task, error) {
	task, err := s.orchestratorAdapter.GetTask()
	if err != nil {
		if errors.Is(err, errs.ErrTaskNotFound) {
			return models.Task{TaskID: 0}, err
		}
		return models.Task{}, err
	}
	return task, nil
}

// PostResult отправляет результат вычисления оркестратору
func (s *AgentService) PostResult(result models.Result) error {
	_, err := s.orchestratorAdapter.ResultTask(result.ExpressionID, result.TaskID, result.Result)
	if err != nil {
		if errors.Is(err, errs.ErrTaskNotFound) {
			return err
		}
		return err
	}
	return nil
}
