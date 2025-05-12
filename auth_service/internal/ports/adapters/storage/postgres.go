package storage

import (
	"context"
	"errors"
	"fmt"
	errs "github.com/jaam8/web_calculator/common-lib/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthPostgresAdapter struct {
	pool *pgxpool.Pool
}

func NewAuthPostgresAdapter(pool *pgxpool.Pool) *AuthPostgresAdapter {
	return &AuthPostgresAdapter{
		pool: pool,
	}
}

// RegisterUser inserts a new user into the db and returns the user ID.
func (a *AuthPostgresAdapter) RegisterUser(login, hashPassword string) (string, error) {
	query := `INSERT INTO users.users (login, password_hash) 
			  VALUES ($1, $2) 
			  RETURNING id`
	var id string
	err := a.pool.QueryRow(context.Background(), query, login, hashPassword).Scan(&id)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrTooManyRows):
			return "", errs.ErrUserAlreadyExists
		default:
			return "", fmt.Errorf("failed to register user: %w", err)
		}
	}
	return id, nil
}

func (a *AuthPostgresAdapter) LoginUser(login string) (string, string, error) {
	query := `SELECT id, password_hash FROM users.users WHERE login = $1`
	var id, hash string
	err := a.pool.QueryRow(context.Background(), query, login).Scan(&id, hash)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return "", "", errs.ErrUserNotFound
		default:
			return "", "", fmt.Errorf("failed to login user: %w", err)
		}
	}
	return id, hash, nil
}
