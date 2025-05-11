package handlers

import (
	"errors"
	errs "github.com/jaam8/web_calculator/common-lib/errors"
	"github.com/jaam8/web_calculator/common-lib/gen/orchestrator"
	"github.com/jaam8/web_calculator/common-lib/logger"
	"github.com/jaam8/web_calculator/gateway/internal/http/schemas"
	"github.com/jaam8/web_calculator/gateway/internal/services"
	"github.com/labstack/echo/v4"
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

func (h *OrchestratorHandler) Calculate(c echo.Context) error {
	var request schemas.CalculateRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, schemas.CannotParseExpression)
	}
	calculateRequest := &orchestrator.CalculateRequest{
		Expression: request.Expression,
	}
	response, err := h.orchestratorService.Calculate(calculateRequest)

	switch {
	case err == nil:
		return c.JSON(http.StatusCreated, response)
	case errors.As(err, &errs.ErrInvalidExpression):
		return c.JSON(http.StatusUnprocessableEntity, schemas.CannotParseExpression)
	default:
		logger.GetOrCreateLoggerFromCtx(c.Request().Context()).Error(
			c.Request().Context(),
			"error in orchestrator service",
			zap.Error(err))
		return c.JSON(http.StatusInternalServerError, schemas.InternalServerError)
	}
}

func (h *OrchestratorHandler) Expressions(c echo.Context) error {
	req := &orchestrator.ExpressionsRequest{Id: 1}
	expressions, err := h.orchestratorService.Expressions(req)
	if err != nil {
		logger.GetOrCreateLoggerFromCtx(c.Request().Context()).Error(
			c.Request().Context(),
			"error in orchestrator service",
			zap.Error(err))
		return c.JSON(http.StatusInternalServerError, schemas.InternalServerError)
	}
	return c.JSON(http.StatusOK, expressions)
}

func (h *OrchestratorHandler) ExpressionByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, schemas.CannotParseId)
	}
	req := &orchestrator.ExpressionByIdRequest{
		Id: int64(id),
	}
	expression, err := h.orchestratorService.ExpressionByID(req)
	switch {
	case err == nil:
		return c.JSON(http.StatusOK, expression)
	case errors.As(err, &errs.ErrExpressionNotFound):
		return c.JSON(http.StatusNotFound, schemas.ExpressionNotFound)
	default:
		return c.JSON(http.StatusInternalServerError, schemas.InternalServerError)
	}
}
