package models

import "time"

// URL represents one row in the urls table.
type URL struct {
	ID        int       `db:"id" json:"id"`
	Code      string    `db:"code" json:"code"`
	LongURL   string    `db:"long_url" json:"long_url"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
