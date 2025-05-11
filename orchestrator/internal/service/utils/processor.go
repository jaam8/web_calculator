package utils

import (
	"context"
	"fmt"
	"github.com/jaam8/web_calculator/common-lib/logger"
	"github.com/jaam8/web_calculator/orchestrator/internal/service/types"
	"go.uber.org/zap"
	"strconv"
)

func Process(rpn []string, tm types.TaskManager, em types.ExpressionManager, expressionID int) {
	ctx, _ := logger.New(context.Background())

	var stack []float64
	for _, v := range rpn {
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			stack = append(stack, num)
			continue
		}
		if len(stack) < 2 {
			em.ExpressionError(expressionID)
			return
		}
		arg2 := stack[len(stack)-1]
		arg1 := stack[len(stack)-2]
		stack = stack[:len(stack)-2]

		task := tm.CreateTask(arg1, arg2, v, expressionID)
		logger.GetLoggerFromCtx(ctx).Debug(ctx,
			fmt.Sprintf("created task with id: %d", task.TaskID),
			zap.Int("expressionID", task.ExpressionID),
			zap.Int("taskID", task.TaskID),
			zap.Float64("arg1", task.Arg1),
			zap.Float64("arg2", task.Arg2),
			zap.String("operator", task.Operation))
		em.AddTask(task)
		result := tm.GetResult()
		logger.GetLoggerFromCtx(ctx).Debug(ctx,
			fmt.Sprintf("got result for task with id: %d", task.TaskID),
			zap.Int("expressionID", expressionID),
			zap.Int("taskID", task.TaskID),
			zap.Float64("result", result.Result))
		stack = append(stack, result.Result)
	}

	if len(stack) != 1 {
		em.ExpressionError(expressionID)
		return
	}
	em.ExpressionDone(expressionID, stack[0])
}
