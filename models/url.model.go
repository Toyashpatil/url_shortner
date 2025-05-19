package models

import "time"

// URL represents one row in the urls table.
type URL struct {
	ID        int       `db:"id" json:"id"`
	BaseURL   string    `db:"base_url" json:"base_url"`
	Code      string    `db:"code" json:"code"`
	LongURL   string    `db:"long_url" json:"long_url"`
	UserID    int       `db:"user_id" json:"user_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
