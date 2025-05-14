package models

import "github.com/google/uuid"

type Expression struct {
	UserId       uuid.UUID `db:"user_id"`
	ExpressionID uuid.UUID `json:"id" db:"id"`
	Status       string    `json:"status" db:"status"`
	Result       *float64  `json:"result,omitempty" db:"result"`
}
