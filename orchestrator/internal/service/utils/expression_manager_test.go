package utils

import (
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

	id1, err1 := em.CreateExpression()
	require.NoError(t, err1)

	id2, err2 := em.CreateExpression()
	require.NoError(t, err2)

	require.NotEqual(t, id1, id2)
	expr, exists := em.GetExpression(id1)
	require.True(t, exists)
	require.Equal(t, "pending", expr.Status)
}

func TestExpressionManager_AddTask(t *testing.T) {
	em := NewExpressionManager(durations)
	task := models.Task{
		ExpressionID:  1,
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
	id, _ := em.CreateExpression()

	em.ExpressionDone(id, 42.0)

	expr, exists := em.GetExpression(id)
	require.True(t, exists)
	require.Equal(t, "done", expr.Status)
	require.NotNil(t, expr.Result)
	require.Equal(t, 42.0, *expr.Result)
}

func TestExpressionManager_ExpressionError(t *testing.T) {
	em := NewExpressionManager(durations)
	id, _ := em.CreateExpression()

	em.ExpressionError(id)

	expr, exists := em.GetExpression(id)
	require.True(t, exists)
	require.Equal(t, "invalid expression", expr.Status)
}

func TestExpressionManager_GetExpressions(t *testing.T) {
	em := NewExpressionManager(durations)
	_, _ = em.CreateExpression()
	_, _ = em.CreateExpression()

	all := em.GetExpressions()
	require.Len(t, all, 2)
}
