package api

import (
	"github.com/jaam8/web_calculator/internal/config"
	o "github.com/jaam8/web_calculator/internal/orchestrator"
	"github.com/labstack/echo/v4"
)

var expressionManager = o.NewExpressionManager()
var taskManager = o.NewTaskManager()
var conf = config.Configs

func RunServer() {
	e := echo.New()

	e.Use(LogMiddleware)
	e.Use(CORSMiddleware)

	e.POST("/api/v1/calculate", CalculateHandler)
	e.GET("/api/v1/expressions", ExpressionsHandler)
	e.GET("/api/v1/expressions/:id", ExpressionByIDHandler)
	e.POST("/internal/task", PostTaskHandler)
	e.GET("/internal/task", GetTaskHandler)
	e.Logger.Fatal(e.Start(":" + conf.Port))
}
