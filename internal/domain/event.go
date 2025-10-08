package domain

import "time"

// Event — структура календарного события.
type Event struct {
	ID      string    `json:"id"`
	UserID  int64     `json:"user_id"`
	Date    time.Time `json:"date"`  // дата без времени (UTC)
	Title   string    `json:"event"` // текст события
	Created time.Time `json:"created_at"`
	Updated time.Time `json:"updated_at"`
}
