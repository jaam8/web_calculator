package api

import (
	"github.com/jaam8/web_calculator/internal/logger"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
)

func CORSMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Устанавливаем CORS заголовки
		c.Response().Header().Set("Access-Control-Allow-Origin", "*") // Разрешаем доступ с любого источника
		c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Если это OPTIONS-запрос, возвращаем 204 статус
		if c.Request().Method == http.MethodOptions {
			return c.NoContent(http.StatusNoContent)
		}

		// Для всех остальных запросов передаем обработку дальше
		return next(c)
	}
}

// LogMiddleware добавляет логирование для каждого запроса
func LogMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		log := logger.Log

		// Логируем входящий запрос
		log.Info("Request",
			zap.String("method", c.Request().Method),
			zap.String("path", c.Request().URL.Path),
		)

		// Выполняем следующий обработчик
		err := next(c)

		// Логируем статусный код ответа
		log.Info("Response",
			zap.Int("status", c.Response().Status),
			zap.String("method", c.Request().Method),
			zap.String("path", c.Request().URL.Path),
		)

		return err
	}
}
