package storage

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/oziev02/Calendar-HTTP-Server/internal/app"
	"github.com/oziev02/Calendar-HTTP-Server/internal/domain"
)

// Memory — потокобезопасное хранилище событий в памяти.
type Memory struct {
	mu         sync.RWMutex
	byID       map[string]domain.Event
	byUserDate map[int64]map[string]map[string]struct{} // userID -> dateKey -> eventIDs
}

func NewMemory() *Memory {
	return &Memory{
		byID:       make(map[string]domain.Event),
		byUserDate: make(map[int64]map[string]map[string]struct{}),
	}
}

func dateKey(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

// Create добавляет новое событие.
func (m *Memory) Create(_ context.Context, e domain.Event) (domain.Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.byID[e.ID]; ok {
		return domain.Event{}, errors.New("id already exists")
	}

	m.byID[e.ID] = e

	if _, ok := m.byUserDate[e.UserID]; !ok {
		m.byUserDate[e.UserID] = make(map[string]map[string]struct{})
	}
	dk := dateKey(e.Date)
	if _, ok := m.byUserDate[e.UserID][dk]; !ok {
		m.byUserDate[e.UserID][dk] = make(map[string]struct{})
	}
	m.byUserDate[e.UserID][dk][e.ID] = struct{}{}

	return e, nil
}

// Update обновляет существующее событие.
func (m *Memory) Update(_ context.Context, e domain.Event) (domain.Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	old, ok := m.byID[e.ID]
	if !ok {
		return domain.Event{}, app.ErrNotFound
	}

	// Если изменились пользователь или дата, нужно обновить индексы.
	if old.UserID != e.UserID || !old.Date.Equal(e.Date) {
		oldDK := dateKey(old.Date)
		delete(m.byUserDate[old.UserID][oldDK], old.ID)
		if len(m.byUserDate[old.UserID][oldDK]) == 0 {
			delete(m.byUserDate[old.UserID], oldDK)
		}

		if _, ok := m.byUserDate[e.UserID]; !ok {
			m.byUserDate[e.UserID] = make(map[string]map[string]struct{})
		}
		dk := dateKey(e.Date)
		if _, ok := m.byUserDate[e.UserID][dk]; !ok {
			m.byUserDate[e.UserID][dk] = make(map[string]struct{})
		}
		m.byUserDate[e.UserID][dk][e.ID] = struct{}{}
	}

	m.byID[e.ID] = e
	return e, nil
}

// Delete удаляет событие.
func (m *Memory) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	e, ok := m.byID[id]
	if !ok {
		return app.ErrNotFound
	}

	delete(m.byID, id)
	dk := dateKey(e.Date)
	delete(m.byUserDate[e.UserID][dk], id)

	if len(m.byUserDate[e.UserID][dk]) == 0 {
		delete(m.byUserDate[e.UserID], dk)
	}
	if len(m.byUserDate[e.UserID]) == 0 {
		delete(m.byUserDate, e.UserID)
	}

	return nil
}

// GetByID возвращает событие по ID.
func (m *Memory) GetByID(_ context.Context, id string) (domain.Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	e, ok := m.byID[id]
	if !ok {
		return domain.Event{}, app.ErrNotFound
	}
	return e, nil
}

// ListForDate возвращает события пользователя на день.
func (m *Memory) ListForDate(_ context.Context, userID int64, date time.Time) ([]domain.Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dk := dateKey(date)
	ids := m.byUserDate[userID][dk]
	res := make([]domain.Event, 0, len(ids))

	for id := range ids {
		res = append(res, m.byID[id])
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Created.Before(res[j].Created)
	})

	return res, nil
}

// ListForRange возвращает события пользователя в диапазоне дат.
func (m *Memory) ListForRange(_ context.Context, userID int64, from, to time.Time) ([]domain.Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]domain.Event, 0)
	for d := from.UTC(); d.Before(to.UTC()); d = d.AddDate(0, 0, 1) {
		dk := dateKey(d)
		for id := range m.byUserDate[userID][dk] {
			res = append(res, m.byID[id])
		}
	}

	sort.Slice(res, func(i, j int) bool {
		if res[i].Date.Equal(res[j].Date) {
			return res[i].Created.Before(res[j].Created)
		}
		return res[i].Date.Before(res[j].Date)
	})

	return res, nil
}

// ExistsByUserDateTitle проверяет, существует ли событие с таким user/date/title.
func (m *Memory) ExistsByUserDateTitle(_ context.Context, userID int64, date time.Time, title string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dk := dateKey(date)
	for id := range m.byUserDate[userID][dk] {
		if m.byID[id].Title == title {
			return true, nil
		}
	}
	return false, nil
}
