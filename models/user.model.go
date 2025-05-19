package models

import "time"

// User represents one row in the users table.
type User struct {
	ID           int       `db:"id"            json:"id"`
	Email        string    `db:"email"         json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"` // omit from JSON responses
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
}
