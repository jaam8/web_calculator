package orchestrator

import (
	"testing"
)

func DoTask(task Task) float64 {
	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2
	case "-":
		return task.Arg1 - task.Arg2
	case "*":
		return task.Arg1 * task.Arg2
	case "/":
		return task.Arg1 / task.Arg2
	}
	return 0
}

func TestProcess_ValidExpression(t *testing.T) {
	tm := NewTaskManager()
	em := NewExpressionManager()
	expressionID, _ := em.CreateExpression()

	go func() {
		for task := range em.GetTasks() {
			result := Result{TaskID: task.TaskID, Result: DoTask(task)}
			tm.AddResult(result)
		}
	}()

	rpn := []string{"3", "4", "2", "*", "1", "5", "-", "/", "+"}
	Process(rpn, tm, em, expressionID)

	expr, exists := em.GetExpression(expressionID)
	if !exists {
		t.Fatalf("Expression with TaskID %d does not exist", expressionID)
	}
	if expr.Status != "done" {
		t.Errorf("Expected status 'done', got %s", expr.Status)
	}
	if *expr.Result != 1 {
		t.Errorf("Expected result 1, got %f", *expr.Result)
	}
}

func TestProcess_EmptyExpression(t *testing.T) {
	tm := NewTaskManager()
	em := NewExpressionManager()
	expressionID, _ := em.CreateExpression()

	go func() {
		for task := range em.GetTasks() {
			result := Result{TaskID: task.TaskID, Result: DoTask(task)}
			tm.AddResult(result)
		}
	}()

	rpn := []string{}
	Process(rpn, tm, em, expressionID)

	expr, exists := em.GetExpression(expressionID)
	if !exists {
		t.Fatalf("Expression with TaskID %d does not exist", expressionID)
	}
	if expr.Status != "invalid expression" {
		t.Errorf("Expected status 'invalid expression', got %s", expr.Status)
	}
	if expr.Result != nil {
		t.Errorf("Expected result to be nil, got %f", *expr.Result)
	}
}
