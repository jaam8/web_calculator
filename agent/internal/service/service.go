package service

import (
	"context"
	"errors"
	"github.com/jaam8/web_calculator/agent/internal/models"
	"github.com/jaam8/web_calculator/agent/internal/ports"
	errs "github.com/jaam8/web_calculator/common-lib/errors"
	"github.com/jaam8/web_calculator/common-lib/logger"
	"go.uber.org/zap"
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
func (s *AgentService) Work(ctx context.Context, waitTime int) {
	sleepTime := time.Duration(waitTime) * time.Millisecond
	for {
		task, err := s.GetTask()
		for err != nil || task.TaskID == 0 {
			switch {
			case errors.Is(err, errs.ErrTaskNotFound):
				logger.GetLoggerFromCtx(ctx).Warn(ctx,
					"Task not found",
					zap.Error(err))
				time.Sleep(sleepTime)
			case errors.Is(err, errs.ErrInternalServerError):
				logger.GetLoggerFromCtx(ctx).Error(ctx,
					"Internal server error",
					zap.Int("task_id", task.TaskID),
					zap.Error(err))
				time.Sleep(sleepTime)
			default:
				logger.GetLoggerFromCtx(ctx).Error(ctx,
					"Unknown error",
					zap.Int("task_id", task.TaskID),
					zap.Error(err),
				)
			}
			task, err = s.GetTask()
		}

		logger.GetLoggerFromCtx(ctx).Debug(ctx,
			"GOT TASK",
			zap.String("expression_id", task.ExpressionID),
			zap.Int("task_id", task.TaskID),
			zap.String("operation", task.Operation),
			zap.Float64("arg1", task.Arg1),
			zap.Float64("arg2", task.Arg2),
			zap.Duration("operation_time", task.OperationTime),
		)

		result, err := DoTask(task)
		if err != nil {
			switch {
			case errors.Is(err, errs.ErrDivideByZero):
				logger.GetLoggerFromCtx(ctx).Error(ctx,
					"Division by zero error",
					zap.Int("task_id", task.TaskID),
					zap.Error(err),
				)
				continue
			case errors.Is(err, errs.ErrInvalidExpression):
				logger.GetLoggerFromCtx(ctx).Error(ctx,
					"Invalid expression error",
					zap.Int("task_id", task.TaskID),
					zap.Error(err))
				continue
			default:
				logger.GetLoggerFromCtx(ctx).Error(ctx,
					"Unknown error",
					zap.Int("task_id", task.TaskID),
					zap.Error(err),
				)
				continue
			}
		}

		Result := models.Result{
			ExpressionID: task.ExpressionID,
			TaskID:       task.TaskID,
			Result:       result,
		}

		err = s.ResultTask(Result)
		if err != nil {
			logger.GetLoggerFromCtx(ctx).Error(ctx,
				"Error send result for task",
				zap.String("expression_id", Result.ExpressionID),
				zap.Int("task_id", Result.TaskID),
				zap.Float64("result", Result.Result),
				zap.Error(err))
		}
		logger.GetLoggerFromCtx(ctx).Info(ctx,
			"Send result for task",
			zap.String("expression_id", Result.ExpressionID),
			zap.Int("task_id", Result.TaskID),
			zap.Float64("result", Result.Result),
		)
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

// ResultTask отправляет результат вычисления оркестратору
func (s *AgentService) ResultTask(result models.Result) error {
	_, err := s.orchestratorAdapter.ResultTask(result.ExpressionID, result.TaskID, result.Result)
	if err != nil {
		if errors.Is(err, errs.ErrTaskNotFound) {
			return err
		}
		return err
	}
	return nil
}
