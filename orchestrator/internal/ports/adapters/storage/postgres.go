package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	errs "github.com/jaam8/web_calculator/common-lib/errors"
	"github.com/jaam8/web_calculator/orchestrator/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAdapter struct {
	pool *pgxpool.Pool
}

func NewPostgresAdapter(pool *pgxpool.Pool) *PostgresAdapter {
	return &PostgresAdapter{
		pool: pool,
	}
}

func (a *PostgresAdapter) SaveExpression(expression models.Expression) (uuid.UUID, error) {
	query := `INSERT INTO expressions.expressions (user_id, status, result) 
			  VALUES ($1, $2, $3)
			  RETURNING id`
	var id uuid.UUID
	err := a.pool.QueryRow(context.Background(), query,
		expression.UserId,
		expression.Status,
		expression.Result).Scan(&id)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to save expression: %w", err)
	}
	return id, nil
}

func (a *PostgresAdapter) GetExpressionById(userId, id uuid.UUID) (*models.Expression, error) {
	query := `SELECT status, result FROM expressions.expressions
			  WHERE user_id = $1 AND id = $2`
	expr := &models.Expression{UserId: userId, ExpressionID: id}
	err := a.pool.QueryRow(context.Background(), query, userId, id).Scan(expr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrExpressionNotFound
		}
		return nil, fmt.Errorf("failed to get expression: %w", err)
	}
	return expr, nil
}

func (a *PostgresAdapter) GetExpressions(userId uuid.UUID) ([]*models.Expression, error) {
	query := `SELECT id, status, result FROM expressions.expressions
			  WHERE user_id = $1`
	var expressions []*models.Expression
	rows, err := a.pool.Query(context.Background(), query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get expressions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		expr := new(models.Expression)
		err = rows.Scan(&expr.ExpressionID, &expr.Status, &expr.Result)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expression: %w", err)
		}
		expr.UserId = userId
		expressions = append(expressions, expr)
	}

	if len(expressions) == 0 {
		return nil, errs.ErrExpressionNotFound
	}

	return expressions, nil
}

func (a *PostgresAdapter) UpdateExpression(userId, id uuid.UUID, status *string, result *float64) error {
	query := `UPDATE expressions.expressions`
	args := make([]any, 0, 4)
	if status != nil {
		args = append(args, status)
		query += fmt.Sprintf(` SET status = $%d`, len(args))
	}
	if result != nil {
		args = append(args, result)
		query += fmt.Sprintf(` SET result = $%d`, len(args))
	}
	args = append(args, userId, id)
	query += fmt.Sprintf(` WHERE user_id = $%d AND id = $%d`, len(args)-1, len(args))
	_, err := a.pool.Exec(context.Background(), query, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrExpressionNotFound
		}
		return fmt.Errorf("failed to update expression: %w", err)
	}
	return nil
}
