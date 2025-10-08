package app

import "errors"

var (
	ErrNotFound  = errors.New("event not found")
	ErrDuplicate = errors.New("duplicate event (same user, date, and title)")
)
