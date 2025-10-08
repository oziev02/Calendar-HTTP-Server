package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oziev02/Calendar-HTTP-Server/internal/app"
	"github.com/oziev02/Calendar-HTTP-Server/internal/storage"
)

func TestCreateAndFetch(t *testing.T) {
	repo := storage.NewMemory()
	svc := app.NewService(repo)
	ctx := context.Background()
	day := time.Date(2025, 10, 8, 0, 0, 0, 0, time.UTC)

	e1, err := svc.CreateEvent(ctx, 1, day, "Meet")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if e1.ID == "" {
		t.Fatal("id empty")
	}

	list, err := svc.EventsForDay(ctx, 1, day)
	if err != nil {
		t.Fatalf("list day: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("want 1, got %d", len(list))
	}

	week, err := svc.EventsForWeek(ctx, 1, day)
	if err != nil || len(week) != 1 {
		t.Fatalf("week err=%v len=%d", err, len(week))
	}

	month, err := svc.EventsForMonth(ctx, 1, day)
	if err != nil || len(month) != 1 {
		t.Fatalf("month err=%v len=%d", err, len(month))
	}
}

func TestDeleteNotFound(t *testing.T) {
	repo := storage.NewMemory()
	svc := app.NewService(repo)
	err := svc.DeleteEvent(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, app.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestDuplicate(t *testing.T) {
	repo := storage.NewMemory()
	svc := app.NewService(repo)
	ctx := context.Background()
	day := time.Date(2025, 10, 8, 0, 0, 0, 0, time.UTC)
	_, _ = svc.CreateEvent(ctx, 42, day, "Standup")

	_, err := svc.CreateEvent(ctx, 42, day, "Standup")
	if !errors.Is(err, app.ErrDuplicate) {
		t.Fatalf("want ErrDuplicate, got %v", err)
	}
}
