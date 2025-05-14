package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	errs "github.com/jaam8/web_calculator/common-lib/errors"
	"github.com/jaam8/web_calculator/common-lib/gen/orchestrator"
	"github.com/jaam8/web_calculator/common-lib/logger"
	"github.com/jaam8/web_calculator/orchestrator/internal/models"
	"github.com/jaam8/web_calculator/orchestrator/internal/ports"
	"github.com/jaam8/web_calculator/orchestrator/internal/service/helper"
	"github.com/jaam8/web_calculator/orchestrator/internal/service/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"strconv"
)

type OrchestratorService struct {
	orchestrator.OrchestratorServiceServer
	expressionManager types.ExpressionManager
	storage           ports.StorageAdapter
}

func NewOrchestratorService(storage ports.StorageAdapter, expressionManager types.ExpressionManager) *OrchestratorService {
	return &OrchestratorService{
		expressionManager: expressionManager,
		storage:           storage,
	}
}

func (s *OrchestratorService) Calculate(
	ctx context.Context, request *orchestrator.CalculateRequest,
) (*orchestrator.CalculateResponse, error) {
	userId, err := uuid.Parse(request.UserId)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to parse user id",
			zap.String("userID", request.UserId),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to parse user id: %w", err)
	}

	rpn, err := helper.ToRPN(request.Expression)
	logger.GetLoggerFromCtx(ctx).Debug(ctx,
		fmt.Sprintf("RPN for expression: %s", request.Expression),
		zap.Any("rpn", rpn),
		zap.Error(err))
	if err != nil {
		return nil, err
	}
	expr := &models.Expression{
		UserId: userId,
		Status: "pending",
		Result: nil,
	}

	expressionId, err := s.storage.SaveExpression(*expr)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to save expression",
			zap.String("userID", expr.UserId.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to save expression: %w", err)
	}

	expr.ExpressionID = expressionId

	err = s.expressionManager.CreateExpression(expr)
	logger.GetLoggerFromCtx(ctx).Debug(ctx,
		fmt.Sprintf("Created expression with id: %s", expr.ExpressionID),
		zap.Error(err),
	)
	if err != nil {
		return nil, err
	}

	taskManager, err := s.expressionManager.GetTaskManager(expr.ExpressionID)
	if err != nil {
		return nil, errs.ErrInternalServerError
	}

	go s.Process(ctx, taskManager, rpn, userId, expr.ExpressionID)

	return &orchestrator.CalculateResponse{Id: expr.ExpressionID.String()}, nil
}

func (s *OrchestratorService) Expressions(
	ctx context.Context, request *orchestrator.ExpressionsRequest,
) (*orchestrator.ExpressionsResponse, error) {
	userId, err := uuid.Parse(request.UserId)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to parse user id",
			zap.String("userID", request.UserId),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to parse user id: %w", err)
	}

	expressions, err := s.storage.GetExpressions(userId)
	if err != nil {
		if errors.Is(err, errs.ErrExpressionNotFound) {
			logger.GetLoggerFromCtx(ctx).Debug(ctx,
				"no expressions found",
				zap.String("userID", userId.String()),
			)
			return &orchestrator.ExpressionsResponse{}, nil
		}
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to get expressions",
			zap.String("userID", userId.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get expressions: %w", err)
	}

	//expressions := s.expressionManager.GetExpressions()
	exprs := make([]*orchestrator.Expression, len(expressions))
	for i, e := range expressions {
		expr := &orchestrator.Expression{
			Id:     e.ExpressionID.String(),
			Status: e.Status,
			Result: e.Result,
		}
		exprs[i] = expr
	}
	logger.GetLoggerFromCtx(ctx).Info(ctx,
		fmt.Sprintf("got %d expressions", len(expressions)),
		zap.String("user_id", userId.String()),
		zap.Int("expressionsCount", len(expressions)))
	return &orchestrator.ExpressionsResponse{Expressions: exprs}, nil
}

func (s *OrchestratorService) ExpressionById(
	ctx context.Context, request *orchestrator.ExpressionByIdRequest,
) (*orchestrator.ExpressionByIdResponse, error) {
	//userId, err := uuid.Parse(request.UserId)
	//if err != nil {
	//	logger.GetLoggerFromCtx(ctx).Error(ctx,
	//		"failed to parse user id",
	//		zap.String("userID", request.UserId),
	//		zap.Error(err),
	//	)
	//	return nil, fmt.Errorf("failed to parse user id: %w", err)
	//}

	expressionId, err := uuid.Parse(request.Id)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to parse expression id",
			zap.String("expressionID", request.Id),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to parse expression id: %w", err)
	}

	expression, ok := s.expressionManager.GetExpression(expressionId)
	if !ok {
		logger.GetLoggerFromCtx(ctx).Warn(ctx,
			fmt.Sprintf("no expression found for id: %s", request.Id),
			zap.String("expressionID", request.Id),
		)
		return nil, errs.ErrExpressionNotFound
	}

	expr := &orchestrator.Expression{
		Id:     expression.ExpressionID.String(),
		Status: expression.Status,
		Result: expression.Result,
	}
	logger.GetLoggerFromCtx(ctx).Info(ctx,
		"got expression by id",
		zap.String("ExpressionID", expr.Id),
		zap.String("status", expr.Status),
		zap.Float64p("result", expr.Result))
	return &orchestrator.ExpressionByIdResponse{Expression: expr}, nil
}

func (s *OrchestratorService) ResultTask(
	ctx context.Context, request *orchestrator.ResultTaskRequest,
) (*orchestrator.ResultTaskResponse, error) {
	expressionId, err := uuid.Parse(request.ExpressionId)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to parse expression id",
			zap.String("expressionID", request.ExpressionId),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to parse expression id: %w", err)
	}
	taskManager, err := s.expressionManager.GetTaskManager(expressionId)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Debug(ctx,
			fmt.Sprintf("no task manager found for expression with id: %d", request.ExpressionId),
			zap.String("expressionID", request.ExpressionId),
		)
		return nil, errs.ErrTaskNotFound
	}

	exprId, err := uuid.Parse(request.ExpressionId)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to parse expression id",
			zap.String("expressionID", request.ExpressionId),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to parse expression id: %w", err)
	}

	result := models.Result{
		ExpressionID: exprId,
		TaskID:       int(request.Id),
		Result:       request.Result,
	}
	taskManager.AddResult(result)
	logger.GetLoggerFromCtx(ctx).Info(ctx,
		fmt.Sprintf("got result for task with id: %d", request.Id),
		zap.String("expressionID", result.ExpressionID.String()),
		zap.Int("taskID", result.TaskID),
		zap.Float64("result", result.Result),
	)
	return &orchestrator.ResultTaskResponse{Status: "task completed"}, nil
}

func (s *OrchestratorService) GetTask(
	ctx context.Context, _ *emptypb.Empty,
) (*orchestrator.GetTaskResponse, error) {
	select {
	case task := <-s.expressionManager.GetTasks():
		if task.TaskID != 0 {
			logger.GetLoggerFromCtx(ctx).Debug(ctx,
				fmt.Sprintf("send task with id: %d", task.TaskID),
				zap.String("expressionID", task.ExpressionID.String()),
				zap.Int("taskID", task.TaskID),
				zap.Float64("arg1", task.Arg1),
				zap.Float64("arg2", task.Arg2),
				zap.String("operation", task.Operation),
				zap.Duration("operationTime", task.OperationTime),
			)
			return &orchestrator.GetTaskResponse{
				Task: &orchestrator.Task{
					ExpressionId:  task.ExpressionID.String(),
					Id:            int64(task.TaskID),
					Arg1:          task.Arg1,
					Arg2:          task.Arg2,
					Operation:     task.Operation,
					OperationTime: durationpb.New(task.OperationTime),
				},
			}, nil
		}
	default:
		logger.GetLoggerFromCtx(ctx).Warn(ctx, "no task found")
		return nil, errs.ErrTaskNotFound
	}
	logger.GetLoggerFromCtx(ctx).Error(ctx,
		"internal server error",
		zap.Error(errs.ErrInternalServerError),
	)
	return nil, errs.ErrInternalServerError
}

func (s *OrchestratorService) Process(ctx context.Context, tm types.TaskManager, rpn []string, userID, expressionID uuid.UUID) {
	var stack []float64
	for _, v := range rpn {
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			stack = append(stack, num)
			continue
		}
		if len(stack) < 2 {
			status := "invalid expression"
			err := s.storage.UpdateExpression(userID, expressionID, &status, nil)
			if err != nil {
				logger.GetLoggerFromCtx(ctx).Error(ctx,
					"failed to update expression",
					zap.String("userID", userID.String()),
					zap.String("expressionID", expressionID.String()),
					zap.Error(err))
				return
			}
			s.expressionManager.ExpressionError(expressionID)
			return
		}
		arg2 := stack[len(stack)-1]
		arg1 := stack[len(stack)-2]
		stack = stack[:len(stack)-2]

		task := tm.CreateTask(arg1, arg2, v, expressionID)
		logger.GetLoggerFromCtx(ctx).Debug(ctx,
			fmt.Sprintf("created task with id: %d", task.TaskID),
			zap.String("expressionID", task.ExpressionID.String()),
			zap.Int("taskID", task.TaskID),
			zap.Float64("arg1", task.Arg1),
			zap.Float64("arg2", task.Arg2),
			zap.String("operator", task.Operation))
		s.expressionManager.AddTask(task)
		result := tm.GetResult()
		logger.GetLoggerFromCtx(ctx).Debug(ctx,
			fmt.Sprintf("got result for task with id: %d", task.TaskID),
			zap.String("expressionID", expressionID.String()),
			zap.Int("taskID", task.TaskID),
			zap.Float64("result", result.Result))
		stack = append(stack, result.Result)
	}

	if len(stack) != 1 {
		status := "invalid expression"
		err := s.storage.UpdateExpression(userID, expressionID, &status, nil)
		if err != nil {
			logger.GetLoggerFromCtx(ctx).Error(ctx,
				"failed to update expression",
				zap.String("userID", userID.String()),
				zap.String("expressionID", expressionID.String()),
				zap.Error(err))
			return
		}
		s.expressionManager.ExpressionError(expressionID)
		return
	}

	result := stack[0]
	status := "done"
	err := s.storage.UpdateExpression(userID, expressionID, &status, &result)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx,
			"failed to update expression",
			zap.String("userID", userID.String()),
			zap.String("expressionID", expressionID.String()),
			zap.Error(err))
		return
	}
	s.expressionManager.ExpressionDone(expressionID, result)
	logger.GetLoggerFromCtx(ctx).Info(ctx,
		fmt.Sprintf("expression with id: %s completed", expressionID),
		zap.String("expressionID", expressionID.String()),
		zap.Float64("result", result),
		zap.String("status", status),
	)
}
