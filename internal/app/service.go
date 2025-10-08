package app

import (
	"context"
	"time"

	"github.com/oziev02/Calendar-HTTP-Server/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, e domain.Event) (domain.Event, error)
	Update(ctx context.Context, e domain.Event) (domain.Event, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (domain.Event, error)
	ListForDate(ctx context.Context, userID int64, date time.Time) ([]domain.Event, error)
	ListForRange(ctx context.Context, userID int64, from, to time.Time) ([]domain.Event, error)
	ExistsByUserDateTitle(ctx context.Context, userID int64, date time.Time, title string) (bool, error)
}

type Service struct {
	repo  Repository
	clock func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, clock: time.Now}
}

func (s *Service) CreateEvent(ctx context.Context, userID int64, date time.Time, title string) (domain.Event, error) {
	date = date.UTC().Truncate(24 * time.Hour)
	dup, err := s.repo.ExistsByUserDateTitle(ctx, userID, date, title)
	if err != nil {
		return domain.Event{}, err
	}
	if dup {
		return domain.Event{}, ErrDuplicate
	}
	e := domain.Event{
		ID:      newID(),
		UserID:  userID,
		Date:    date,
		Title:   title,
		Created: s.clock().UTC(),
		Updated: s.clock().UTC(),
	}
	return s.repo.Create(ctx, e)
}

func (s *Service) UpdateEvent(ctx context.Context, id string, userID *int64, date *time.Time, title *string) (domain.Event, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Event{}, err
	}
	if userID != nil {
		existing.UserID = *userID
	}
	if date != nil {
		existing.Date = date.UTC().Truncate(24 * time.Hour)
	}
	if title != nil {
		existing.Title = *title
	}
	existing.Updated = s.clock().UTC()
	return s.repo.Update(ctx, existing)
}

func (s *Service) DeleteEvent(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) EventsForDay(ctx context.Context, userID int64, date time.Time) ([]domain.Event, error) {
	date = date.UTC().Truncate(24 * time.Hour)
	return s.repo.ListForDate(ctx, userID, date)
}

func (s *Service) EventsForWeek(ctx context.Context, userID int64, date time.Time) ([]domain.Event, error) {
	date = date.UTC().Truncate(24 * time.Hour)
	offset := (int(date.Weekday()) + 6) % 7
	monday := date.AddDate(0, 0, -offset)
	nextMonday := monday.AddDate(0, 0, 7)
	return s.repo.ListForRange(ctx, userID, monday, nextMonday)
}

func (s *Service) EventsForMonth(ctx context.Context, userID int64, date time.Time) ([]domain.Event, error) {
	date = date.UTC().Truncate(24 * time.Hour)
	first := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
	nextMonth := first.AddDate(0, 1, 0)
	return s.repo.ListForRange(ctx, userID, first, nextMonth)
}

// ---- ID генератор ----
var counter = time.Now().UnixNano()

func newID() string {
	counter++
	return time.Now().UTC().Format("20060102T150405Z") + ":" + itoa(counter)
}

func itoa(n int64) string {
	buf := [20]byte{}
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
		if n == 0 {
			break
		}
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
