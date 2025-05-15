package middlewares

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func CORSMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		allowedOrigins := []string{
			"http://localhost:8081", // фронтенд
			"http://frontend:8081",  // внутри Docker
		}
		origin := c.Request().Header.Get("Origin")
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				c.Response().Header().Set("Access-Control-Allow-Origin", allowed)
				break
			}
		}

		//c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Response().Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request().Method == http.MethodOptions {
			return c.NoContent(http.StatusNoContent)
		}

		return next(c)
	}
}
