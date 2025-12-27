package models

import "time"

type User struct {
	ID           int64     `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Username     string    `db:"username" json:"username"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
	IsActive     bool      `db:"is_active" json:"is_active"`
	Role         string    `db:"role" json:"role"`
}
