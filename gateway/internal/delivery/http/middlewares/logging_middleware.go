package middlewares

import (
	"context"
	"github.com/jaam8/web_calculator/common-lib/logger"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// LogMiddleware добавляет логирование для каждого запроса
func LogMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, err := logger.New(c.Request().Context())
		if err != nil {
			ctx, _ = logger.New(context.Background())

		}

		c.SetRequest(c.Request().WithContext(ctx))

		logger.GetOrCreateLoggerFromCtx(ctx).Info(
			ctx,
			"Request",
			zap.String("method", c.Request().Method),
			zap.String("path", c.Request().URL.Path),
		)
		err = next(c)

		logger.GetOrCreateLoggerFromCtx(ctx).Info(
			ctx,
			"Response",
			zap.Int("status", c.Response().Status),
			zap.String("method", c.Request().Method),
			zap.String("path", c.Request().URL.Path),
		)

		return err
	}
}
