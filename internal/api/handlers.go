package api

import (
	o "github.com/jaam8/web_calculator/internal/orchestrator"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type ExpressionRequest struct {
	Value string `json:"expression"`
}

func CalculateHandler(c echo.Context) error {
	var expression ExpressionRequest
	if err := c.Bind(&expression); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, "cannot parse request")
	}
	rpn, err := o.RPN(expression.Value)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, "invalid expression")
	}
	expressionID, err := expressionManager.CreateExpression()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "cannot create expression")
	}
	taskManager, exists := expressionManager.GetTaskManager(expressionID)
	if !exists {
		return c.JSON(http.StatusInternalServerError, "cannot get task manager")
	}
	go o.Process(rpn, taskManager, expressionManager, expressionID)
	return c.JSON(http.StatusCreated, map[string]int{"id": expressionID})
}

func ExpressionsHandler(c echo.Context) error {
	expressions := expressionManager.GetExpressions()
	return c.JSON(http.StatusOK, map[string][]*o.Expression{"expressions": expressions})
}

func ExpressionByIDHandler(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "cannot parse id")
	}
	expression, ok := expressionManager.GetExpression(id)
	switch ok {
	case true:
		return c.JSON(http.StatusOK, expression)
	case false:
		return c.JSON(http.StatusNotFound, "expression not found")
	default:
		return c.JSON(http.StatusInternalServerError, "internal server error")
	}
}

func PostTaskHandler(c echo.Context) error {
	var result o.Result
	if err := c.Bind(&result); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, "cannot parse request")
	}
	taskManager, ok := expressionManager.GetTaskManager(result.ExpressionID)
	switch ok {
	case true:
		taskManager.AddResult(result)
		return c.JSON(http.StatusOK, "task completed")
	case false:
		return c.JSON(http.StatusNotFound, "task not found")
	default:
		return c.JSON(http.StatusInternalServerError, "internal server error")
	}
}

func GetTaskHandler(c echo.Context) error {
	select {
	case task := <-expressionManager.GetTasks():
		if task.TaskID != 0 {
			return c.JSON(http.StatusOK, task)
		}
	default:
		return c.JSON(http.StatusNotFound, "task not found")
	}
	return c.JSON(http.StatusInternalServerError, "internal server error")
}
