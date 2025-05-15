package utils

import (
	"github.com/google/uuid"
	"github.com/jaam8/web_calculator/orchestrator/internal/models"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var durations = map[string]int{
	"+": 100,
	"-": 100,
	"*": 100,
	"/": 100,
}

func TestExpressionManager_CreateExpression(t *testing.T) {
	em := NewExpressionManager(durations)

	// Создаем тестовое выражение
	expr1 := &models.Expression{
		UserId: uuid.New(),
		Status: "pending",
		Result: nil,
	}
	expr1.ExpressionID = uuid.New()

	err1 := em.CreateExpression(expr1)
	require.NoError(t, err1)

	// Создаем второе тестовое выражение
	expr2 := &models.Expression{
		UserId: uuid.New(),
		Status: "pending",
		Result: nil,
	}
	expr2.ExpressionID = uuid.New()

	err2 := em.CreateExpression(expr2)
	require.NoError(t, err2)

	// Проверяем, что идентификаторы разные
	require.NotEqual(t, expr1.ExpressionID, expr2.ExpressionID)

	// Проверяем, что выражение сохранено
	gotExpr, exists := em.GetExpression(expr1.ExpressionID)
	require.True(t, exists)
	require.Equal(t, "pending", gotExpr.Status)
}

func TestExpressionManager_AddTask(t *testing.T) {
	em := NewExpressionManager(durations)
	expressionID := uuid.New()

	task := models.Task{
		ExpressionID:  expressionID,
		TaskID:        1,
		Arg1:          2,
		Arg2:          2,
		Operation:     "+",
		OperationTime: time.Millisecond * time.Duration(durations["+"]),
	}

	em.AddTask(task)

	select {
	case got := <-em.GetTasks():
		require.Equal(t, task, got)
	case <-time.After(time.Second):
		t.Fatal("did not receive task in time")
	}
}

func TestExpressionManager_ExpressionDone(t *testing.T) {
	em := NewExpressionManager(durations)

	// Создаем тестовое выражение
	expressionID := uuid.New()
	expr := &models.Expression{
		ExpressionID: expressionID,
		UserId:       uuid.New(),
		Status:       "pending",
		Result:       nil,
	}

	_ = em.CreateExpression(expr)

	em.ExpressionDone(expressionID, 42.0)

	gotExpr, exists := em.GetExpression(expressionID)
	require.True(t, exists)
	require.Equal(t, "done", gotExpr.Status)
	require.NotNil(t, gotExpr.Result)
	require.Equal(t, 42.0, *gotExpr.Result)
}

func TestExpressionManager_ExpressionError(t *testing.T) {
	em := NewExpressionManager(durations)

	// Создаем тестовое выражение
	expressionID := uuid.New()
	expr := &models.Expression{
		ExpressionID: expressionID,
		UserId:       uuid.New(),
		Status:       "pending",
		Result:       nil,
	}

	_ = em.CreateExpression(expr)

	em.ExpressionError(expressionID)

	gotExpr, exists := em.GetExpression(expressionID)
	require.True(t, exists)
	require.Equal(t, "invalid expression", gotExpr.Status)
}

func TestExpressionManager_GetExpressions(t *testing.T) {
	em := NewExpressionManager(durations)

	// Создаем два тестовых выражения
	expr1 := &models.Expression{
		ExpressionID: uuid.New(),
		UserId:       uuid.New(),
		Status:       "pending",
		Result:       nil,
	}

	expr2 := &models.Expression{
		ExpressionID: uuid.New(),
		UserId:       uuid.New(),
		Status:       "pending",
		Result:       nil,
	}

	_ = em.CreateExpression(expr1)
	_ = em.CreateExpression(expr2)

	all := em.GetExpressions()
	require.Len(t, all, 2)
}

func TestExpressionManager_GetTaskManager(t *testing.T) {
	em := NewExpressionManager(durations)

	// Создаем тестовое выражение
	expressionID := uuid.New()
	expr := &models.Expression{
		ExpressionID: expressionID,
		UserId:       uuid.New(),
		Status:       "pending",
		Result:       nil,
	}

	_ = em.CreateExpression(expr)

	// Проверяем, что TaskManager создан для выражения
	taskManager, err := em.GetTaskManager(expressionID)
	require.NoError(t, err)
	require.NotNil(t, taskManager)

	// Проверяем несуществующий TaskManager
	invalidID := uuid.New()
	_, err = em.GetTaskManager(invalidID)
	require.Error(t, err)
}
