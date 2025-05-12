package middlewares

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func CORSMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if c.Request().Method == http.MethodOptions {
			return c.NoContent(http.StatusNoContent)
		}

		return next(c)
	}
}
