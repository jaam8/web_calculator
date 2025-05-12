package handlers

import (
	"errors"
	errs "github.com/jaam8/web_calculator/common-lib/errors"
	"github.com/jaam8/web_calculator/common-lib/gen/orchestrator"
	"github.com/jaam8/web_calculator/common-lib/logger"
	"github.com/jaam8/web_calculator/gateway/internal/http/schemas"
	"github.com/jaam8/web_calculator/gateway/internal/services"
	"github.com/labstack/echo/v4"
	_ "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type OrchestratorHandler struct {
	orchestratorService *services.OrchestratorService
}

func NewOrchestratorHandler(orchestratorService *services.OrchestratorService) *OrchestratorHandler {
	return &OrchestratorHandler{
		orchestratorService: orchestratorService,
	}
}

// @Summary Calculate mathematical expression
// @Description Evaluates a mathematical expression and returns the result
// @Tags Orchestrator
// @Accept json
// @Produce json
// @Param expression body schemas.CalculateRequest true "Expression to calculate"
// @Success 201 {object} schemas.CalculateResponse
// @Failure 422 {object} schemas.CannotParseExpression
// @Failure 500 {object} schemas.InternalServerError
// @Router /calculate [post]
func (h *OrchestratorHandler) Calculate(c echo.Context) error {
	var request schemas.CalculateRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, schemas.CannotParseExpressionMsg)
	}
	calculateRequest := &orchestrator.CalculateRequest{
		Expression: request.Expression,
	}
	response, err := h.orchestratorService.Calculate(calculateRequest)

	switch {
	case err == nil:
		return c.JSON(http.StatusCreated, response)
	case errors.As(err, &errs.ErrInvalidExpression):
		return c.JSON(http.StatusUnprocessableEntity, schemas.CannotParseExpressionMsg)
	default:
		logger.GetOrCreateLoggerFromCtx(c.Request().Context()).Error(
			c.Request().Context(),
			"error in orchestrator service",
			zap.Error(err))
		return c.JSON(http.StatusInternalServerError, schemas.InternalServerErrorMsg)
	}
}

// @Summary Get all expressions
// @Description Returns a list of all calculated expressions
// @Tags Orchestrator
// @Produce json
// @Success 200 {object} schemas.ExpressionsResponse
// @Failure 500 {object} schemas.InternalServerError
// @Router /expressions [get]
func (h *OrchestratorHandler) Expressions(c echo.Context) error {
	req := &orchestrator.ExpressionsRequest{Id: 1}
	expressions, err := h.orchestratorService.Expressions(req)
	if err != nil {
		logger.GetOrCreateLoggerFromCtx(c.Request().Context()).Error(
			c.Request().Context(),
			"error in orchestrator service",
			zap.Error(err))
		return c.JSON(http.StatusInternalServerError, schemas.InternalServerErrorMsg)
	}
	return c.JSON(http.StatusOK, expressions)
}

// @Summary Get expression by ID
// @Description Returns a specific expression by its ID
// @Tags Orchestrator
// @Produce json
// @Param id path int true "Expression ID"
// @Success 200 {object} schemas.ExpressionByIdResponse
// @Failure 404 {object} schemas.ExpressionNotFound
// @Failure 500 {object} schemas.InternalServerError
// @Router /expressions/{id} [get]
func (h *OrchestratorHandler) ExpressionByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, schemas.CannotParseIdMsg)
	}
	req := &orchestrator.ExpressionByIdRequest{
		Id: int64(id),
	}
	expression, err := h.orchestratorService.ExpressionByID(req)
	switch {
	case err == nil:
		return c.JSON(http.StatusOK, expression)
	case errors.As(err, &errs.ErrExpressionNotFound):
		return c.JSON(http.StatusNotFound, schemas.ExpressionNotFoundMsg)
	default:
		return c.JSON(http.StatusInternalServerError, schemas.InternalServerErrorMsg)
	}
}
