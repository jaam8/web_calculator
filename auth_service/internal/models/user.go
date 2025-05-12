package models

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID
	Login        string `json:"login" db:"login"`
	PasswordHash string `json:"password_hash" db:"password_hash"`
}
