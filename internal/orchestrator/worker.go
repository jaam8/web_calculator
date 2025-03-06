package orchestrator

import (
	"github.com/jaam8/web_calculator/internal/config"
	"strconv"
)

var conf = config.Configs

func Process(rpn []string, tm *TaskManager, em *ExpressionManager, expressionID int) {
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

		task := tm.CreateTask(arg1, arg2, v)
		tm.AddTask(task)
		result := tm.GetResult()
		stack = append(stack, result.Result)
	}

	if len(stack) != 1 {
		em.ExpressionError(expressionID)
		return
	}
	em.ExpressionDone(expressionID, stack[0])
	return
}
