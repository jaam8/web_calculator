package handlers

import (
	"errors"
	errs "github.com/jaam8/web_calculator/common-lib/errors"
	auth "github.com/jaam8/web_calculator/common-lib/gen/auth_service"
	"github.com/jaam8/web_calculator/gateway/internal/delivery/grpc"
	"github.com/jaam8/web_calculator/gateway/internal/delivery/http/schemas"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type AuthServiceHandler struct {
	authService *grpc.AuthService
	AccessTTL   time.Duration
	RefreshTTL  time.Duration
}

func NewAuthServiceHandler(authService *grpc.AuthService,
	AccessTTL, RefreshTTL time.Duration) *AuthServiceHandler {
	return &AuthServiceHandler{
		authService: authService,
	}
}

// @Summary Login user
// @Description Authenticates a user by login and password. Write access and refresh tokens to cookies.
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body schemas.LoginRequest true "Login credentials"
// @Success 200
// @Failure 400 {object} schemas.EmptyLogin "Empty login"
// @Failure 400 {object} schemas.EmptyPassword "Empty password"
// @Failure 401 {object} schemas.WrongCredentials "Wrong credentials"
// @Failure 422 {object} schemas.CannotParseRequest "Cannot parse request"
// @Failure 500 {object} schemas.InternalServerError "Internal server error"
// @Router /login [post]
func (h *AuthServiceHandler) Login(c echo.Context) error {
	var request schemas.LoginRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, schemas.CannotParseRequestMsg)
	}
	loginRequest := &auth.LoginRequest{
		Login:    request.Login,
		Password: request.Password,
	}
	response, err := h.authService.Login(loginRequest)
	switch {
	case err == nil:
		c.SetCookie(&http.Cookie{
			Name:     "access_token",
			Value:    response.AccessToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(h.AccessTTL),
		})
		c.SetCookie(&http.Cookie{
			Name:     "refresh_token",
			Value:    response.RefreshToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
			Expires:  time.Now().Add(h.RefreshTTL),
		})
		c.Set("userID", response.UserId)
		return c.JSON(http.StatusOK, auth.LoginResponse{
			AccessToken:  response.AccessToken,
			RefreshToken: response.RefreshToken,
		})
	case errors.Is(err, errs.ErrWrongPassword):
		return c.JSON(http.StatusUnauthorized, schemas.WrongCredentialsMsg)
	case errors.Is(err, errs.ErrUserNotFound):
		return c.JSON(http.StatusUnauthorized, schemas.WrongCredentialsMsg)
	case errors.Is(err, errs.ErrEmptyLogin):
		return c.JSON(http.StatusBadRequest, schemas.EmptyLoginMsg)
	case errors.Is(err, errs.ErrEmptyPassword):
		return c.JSON(http.StatusBadRequest, schemas.EmptyPasswordMsg)
	default:
		return c.JSON(http.StatusInternalServerError, schemas.InternalServerErrorMsg)
	}
}

// @Summary Register new user
// @Description Registers a new user with login and password. Returns user ID in response body.
// @Tags Auth
// @Accept json
// @Produce json
// @Param register body schemas.RegisterRequest true "Register credentials"
// @Success 200 {object} schemas.RegisterResponse
// @Failure 400 {object} schemas.EmptyLogin "Empty login"
// @Failure 400 {object} schemas.EmptyPassword "Empty password"// @Failure 409 {object} schemas.MessageResponse "User already exists"
// @Failure 401 {object} schemas.WrongCredentials "Wrong credentials"
// @Failure 422 {object} schemas.CannotParseRequest "Cannot parse request"
// @Failure 500 {object} schemas.InternalServerError "Internal server error"
// @Router /register [post]
func (h *AuthServiceHandler) Register(c echo.Context) error {
	var request schemas.RegisterRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, schemas.CannotParseRequestMsg)
	}
	registerRequest := &auth.RegisterRequest{
		Login:    request.Login,
		Password: request.Password,
	}
	response, err := h.authService.Register(registerRequest)
	switch {
	case err == nil:
		return c.JSON(http.StatusOK, schemas.RegisterResponse{UserId: response.UserId})
	case errors.Is(err, errs.ErrUserAlreadyExists):
		return c.JSON(http.StatusUnauthorized, schemas.WrongCredentialsMsg)
	case errors.Is(err, errs.ErrEmptyLogin):
		return c.JSON(http.StatusBadRequest, schemas.EmptyLoginMsg)
	case errors.Is(err, errs.ErrEmptyPassword):
		return c.JSON(http.StatusBadRequest, schemas.EmptyPasswordMsg)
	default:
		return c.JSON(http.StatusInternalServerError, schemas.InternalServerErrorMsg)
	}
}

// @Summary Refresh access and refresh tokens
// @Description Refreshes access and refresh tokens using the refresh token from the cookie. Returns new tokens in cookies.
// @Tags Auth
// @Accept json
// @Produce json
// @Success 204 "Success generating new tokens"
// @Failure 401 {object} schemas.TokenExpired "Token expired"
// @Failure 401 {object} schemas.TokenExpiredOrInvalid "Token expired or invalid"
// @Failure 500 {object} schemas.InternalServerError "Internal server error"
// @Router /refresh [post]
func (h *AuthServiceHandler) Refresh(c echo.Context) error {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, schemas.TokenExpiredOrInvalidMsg)
	}
	refreshRequest := &auth.RefreshRequest{
		RefreshToken: refreshToken.Value,
	}
	response, err := h.authService.Refresh(refreshRequest)
	switch {
	case err == nil:
		c.SetCookie(&http.Cookie{
			Name:     "access_token",
			Value:    response.AccessToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(h.AccessTTL),
		})
		c.SetCookie(&http.Cookie{
			Name:     "refresh_token",
			Value:    response.RefreshToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
			Expires:  time.Now().Add(h.RefreshTTL),
		})

		return c.NoContent(http.StatusNoContent)
	case errors.Is(err, errs.ErrTokenExpired):
		return c.JSON(http.StatusUnauthorized, schemas.TokenExpiredMsg)
	case errors.Is(err, errs.ErrInvalidToken):
		return c.JSON(http.StatusUnauthorized, schemas.TokenExpiredOrInvalidMsg)
	default:
		return c.JSON(http.StatusInternalServerError, schemas.InternalServerErrorMsg)
	}
}
