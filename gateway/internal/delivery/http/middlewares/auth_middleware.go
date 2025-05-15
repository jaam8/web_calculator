package middlewares

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	errs "github.com/jaam8/web_calculator/common-lib/errors"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"time"
)

func AuthMiddleware(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var accessToken string
			auth := c.Request().Header.Get("Authorization")

			if auth == "Bearer undefined" || auth == "" {
				authAccessCookie, _ := c.Cookie("access_token")
				if authAccessCookie == nil {
					return echo.NewHTTPError(http.StatusUnauthorized, errs.ErrInvalidToken)
				}
				accessToken = authAccessCookie.Value
			} else {
				parts := strings.SplitN(auth, " ", 2)
				if len(parts) != 2 || parts[0] != "Bearer" {
					return echo.NewHTTPError(http.StatusUnauthorized, errs.ErrInvalidToken)
				}
				accessToken = parts[1]
			}

			sub, isRefresh, expTime, err := ParseJwt(accessToken, jwtSecret)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, err)
			}

			if isRefresh {
				return echo.NewHTTPError(http.StatusUnauthorized, errs.ErrInvalidToken)
			}

			if time.Now().After(expTime) {
				return echo.NewHTTPError(http.StatusUnauthorized, errs.ErrTokenExpired)
			}

			c.Set("userID", sub)

			return next(c)
		}
	}
}

func ParseJwt(rawToken, jwtSecret string) (string, bool, time.Time, error) {
	token, err := jwt.Parse(rawToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return "", false, time.Time{}, errs.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", false, time.Time{}, errs.ErrInvalidToken
	}

	sub, err := claims.GetSubject()
	if err != nil || sub == "" {
		return "", false, time.Time{}, errs.ErrInvalidToken
	}

	exp, err := token.Claims.GetExpirationTime()
	if err != nil {
		return "", false, time.Time{}, err
	}

	expTime := time.Unix(exp.Unix(), 0)

	isRefresh, _ := claims["is_refresh"].(bool)

	return sub, isRefresh, expTime, nil
}
