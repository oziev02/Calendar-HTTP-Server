package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/oziev02/Calendar-HTTP-Server/internal/app"
	"github.com/oziev02/Calendar-HTTP-Server/internal/domain"
	"github.com/oziev02/Calendar-HTTP-Server/pkg/jsonx"
)

type Service interface {
	CreateEvent(ctx context.Context, userID int64, date time.Time, title string) (domain.Event, error)
	UpdateEvent(ctx context.Context, id string, userID *int64, date *time.Time, title *string) (domain.Event, error)
	DeleteEvent(ctx context.Context, id string) error
	EventsForDay(ctx context.Context, userID int64, date time.Time) ([]domain.Event, error)
	EventsForWeek(ctx context.Context, userID int64, date time.Time) ([]domain.Event, error)
	EventsForMonth(ctx context.Context, userID int64, date time.Time) ([]domain.Event, error)
}

type Handlers struct{ Svc Service }

const dateLayout = "2006-01-02"

func parseDate(s string) (time.Time, error) { return time.Parse(dateLayout, s) }

// parseBody умеет читать как JSON, так и x-www-form-urlencoded.
func parseBody(r *http.Request, dst any) error {
	ct := r.Header.Get("Content-Type")
	if ct == "" || strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err != nil {
			return err
		}
		data := make(map[string]any, len(r.Form))
		for k := range r.Form {
			data[k] = r.Form.Get(k)
		}
		b, _ := json.Marshal(data)
		return json.Unmarshal(b, dst)
	}
	// по умолчанию — JSON
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.Unmarshal(body, dst)
}

func (h *Handlers) CreateEvent(w http.ResponseWriter, r *http.Request) {
	type req struct {
		UserID string `json:"user_id"`
		Date   string `json:"date"`
		Title  string `json:"event"`
	}
	var in req
	if err := parseBody(r, &in); err != nil {
		jsonx.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	uid, err := strconv.ParseInt(in.UserID, 10, 64)
	if err != nil {
		jsonx.WriteError(w, http.StatusBadRequest, "invalid user_id")
		return
	}
	d, err := parseDate(in.Date)
	if err != nil {
		jsonx.WriteError(w, http.StatusBadRequest, "invalid date (YYYY-MM-DD)")
		return
	}
	if strings.TrimSpace(in.Title) == "" {
		jsonx.WriteError(w, http.StatusBadRequest, "event is required")
		return
	}
	e, err := h.Svc.CreateEvent(r.Context(), uid, d, in.Title)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, app.ErrDuplicate) {
			status = http.StatusServiceUnavailable
		}
		jsonx.WriteError(w, status, err.Error())
		return
	}
	jsonx.WriteOK(w, e)
}

func (h *Handlers) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	type req struct {
		ID     string  `json:"id"`
		UserID *string `json:"user_id"`
		Date   *string `json:"date"`
		Title  *string `json:"event"`
	}
	var in req
	if err := parseBody(r, &in); err != nil {
		jsonx.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if strings.TrimSpace(in.ID) == "" {
		jsonx.WriteError(w, http.StatusBadRequest, "id is required")
		return
	}

	var uid *int64
	if in.UserID != nil {
		u, err := strconv.ParseInt(*in.UserID, 10, 64)
		if err != nil {
			jsonx.WriteError(w, http.StatusBadRequest, "invalid user_id")
			return
		}
		uid = &u
	}

	var dt *time.Time
	if in.Date != nil {
		d, err := parseDate(*in.Date)
		if err != nil {
			jsonx.WriteError(w, http.StatusBadRequest, "invalid date (YYYY-MM-DD)")
			return
		}
		dt = &d
	}

	e, err := h.Svc.UpdateEvent(r.Context(), in.ID, uid, dt, in.Title)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, app.ErrNotFound) {
			status = http.StatusServiceUnavailable
		}
		jsonx.WriteError(w, status, err.Error())
		return
	}
	jsonx.WriteOK(w, e)
}

func (h *Handlers) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	type req struct {
		ID string `json:"id"`
	}
	var in req
	if err := parseBody(r, &in); err != nil {
		jsonx.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if strings.TrimSpace(in.ID) == "" {
		jsonx.WriteError(w, http.StatusBadRequest, "id is required")
		return
	}
	if err := h.Svc.DeleteEvent(r.Context(), in.ID); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, app.ErrNotFound) {
			status = http.StatusServiceUnavailable
		}
		jsonx.WriteError(w, status, err.Error())
		return
	}
	jsonx.WriteOK(w, "deleted")
}

func (h *Handlers) EventsForDay(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	dateStr := r.URL.Query().Get("date")
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		jsonx.WriteError(w, http.StatusBadRequest, "invalid user_id")
		return
	}
	d, err := parseDate(dateStr)
	if err != nil {
		jsonx.WriteError(w, http.StatusBadRequest, "invalid date (YYYY-MM-DD)")
		return
	}
	res, err := h.Svc.EventsForDay(r.Context(), uid, d)
	if err != nil {
		jsonx.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonx.WriteOK(w, res)
}

func (h *Handlers) EventsForWeek(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	dateStr := r.URL.Query().Get("date")
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		jsonx.WriteError(w, http.StatusBadRequest, "invalid user_id")
		return
	}
	d, err := parseDate(dateStr)
	if err != nil {
		jsonx.WriteError(w, http.StatusBadRequest, "invalid date (YYYY-MM-DD)")
		return
	}
	res, err := h.Svc.EventsForWeek(r.Context(), uid, d)
	if err != nil {
		jsonx.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonx.WriteOK(w, res)
}

func (h *Handlers) EventsForMonth(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	dateStr := r.URL.Query().Get("date")
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		jsonx.WriteError(w, http.StatusBadRequest, "invalid user_id")
		return
	}
	d, err := parseDate(dateStr)
	if err != nil {
		jsonx.WriteError(w, http.StatusBadRequest, "invalid date (YYYY-MM-DD)")
		return
	}
	res, err := h.Svc.EventsForMonth(r.Context(), uid, d)
	if err != nil {
		jsonx.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonx.WriteOK(w, res)
}
