package subscription

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	modelsub "test_task/internal/domain/models/subscription"
	logmid "test_task/internal/middleware/loger_middleware"

	"github.com/google/uuid"
)

type mockUsecase struct {
	createFn func(ctx context.Context, s modelsub.Subscription) (modelsub.Subscription, error)
	getFn    func(ctx context.Context, id uuid.UUID) (modelsub.Subscription, error)
	updateFn func(ctx context.Context, id uuid.UUID, s modelsub.Subscription) (modelsub.Subscription, error)
	deleteFn func(ctx context.Context, id uuid.UUID) error
	listFn   func(ctx context.Context, f modelsub.ListFilter) ([]modelsub.Subscription, int, error)
	sumFn    func(ctx context.Context, f modelsub.SummaryFilter) (int64, error)
}

func (m *mockUsecase) Create(ctx context.Context, s modelsub.Subscription) (modelsub.Subscription, error) {
	return m.createFn(ctx, s)
}
func (m *mockUsecase) GetSub(ctx context.Context, id uuid.UUID) (modelsub.Subscription, error) {
	return m.getFn(ctx, id)
}
func (m *mockUsecase) UpdateSub(ctx context.Context, id uuid.UUID, s modelsub.Subscription) (modelsub.Subscription, error) {
	return m.updateFn(ctx, id, s)
}
func (m *mockUsecase) Delete(ctx context.Context, id uuid.UUID) error { return m.deleteFn(ctx, id) }
func (m *mockUsecase) List(ctx context.Context, f modelsub.ListFilter) ([]modelsub.Subscription, int, error) {
	return m.listFn(ctx, f)
}
func (m *mockUsecase) Summary(ctx context.Context, f modelsub.SummaryFilter) (int64, error) {
	return m.sumFn(ctx, f)
}

func TestCreateSubscription_OK(t *testing.T) {
	now := time.Now().UTC()
	id := uuid.New()
	userID := uuid.New()

	u := &mockUsecase{
		createFn: func(ctx context.Context, s modelsub.Subscription) (modelsub.Subscription, error) {
			if s.ServiceName != "Yandex Plus" {
				t.Fatalf("service name mismatch: %q", s.ServiceName)
			}
			if s.Price != 400 {
				t.Fatalf("price mismatch: %d", s.Price)
			}
			if s.UserID != userID {
				t.Fatalf("userID mismatch: %s", s.UserID)
			}
			if s.StartDate.Year() != 2025 || s.StartDate.Month() != time.July || s.StartDate.Day() != 1 {
				t.Fatalf("start_date mismatch: %v", s.StartDate)
			}

			s.ID = id
			s.CreatedAt = now
			s.UpdatedAt = now
			return s, nil
		},
		getFn: func(context.Context, uuid.UUID) (modelsub.Subscription, error) { return modelsub.Subscription{}, nil },
		updateFn: func(context.Context, uuid.UUID, modelsub.Subscription) (modelsub.Subscription, error) {
			return modelsub.Subscription{}, nil
		},
		deleteFn: func(context.Context, uuid.UUID) error { return nil },
		listFn:   func(context.Context, modelsub.ListFilter) ([]modelsub.Subscription, int, error) { return nil, 0, nil },
		sumFn:    func(context.Context, modelsub.SummaryFilter) (int64, error) { return 0, nil },
	}

	log := logmid.NewLogger("error")
	h := New(log, u)
	r := Router(log, h)

	body := map[string]any{
		"service_name": "Yandex Plus",
		"price":        400,
		"user_id":      userID.String(),
		"start_date":   "07-2025",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/", bytes.NewReader(b))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("want %d, got %d, body=%s", http.StatusCreated, w.Code, w.Body.String())
	}

	var resp modelsub.SubscriptionResp
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.ID != id.String() {
		t.Fatalf("id mismatch: %s", resp.ID)
	}
	if resp.ServiceName != "Yandex Plus" {
		t.Fatalf("service mismatch: %s", resp.ServiceName)
	}
	if resp.Price != 400 {
		t.Fatalf("price mismatch: %d", resp.Price)
	}
	if resp.UserID != userID.String() {
		t.Fatalf("user mismatch: %s", resp.UserID)
	}
	if resp.StartDate != "07-2025" {
		t.Fatalf("start_date mismatch: %s", resp.StartDate)
	}
}

func TestCreateSubscription_BadJSON(t *testing.T) {
	u := &mockUsecase{
		createFn: func(context.Context, modelsub.Subscription) (modelsub.Subscription, error) {
			t.Fatalf("usecase must not be called on bad json")
			return modelsub.Subscription{}, nil
		},
		getFn: func(context.Context, uuid.UUID) (modelsub.Subscription, error) { return modelsub.Subscription{}, nil },
		updateFn: func(context.Context, uuid.UUID, modelsub.Subscription) (modelsub.Subscription, error) {
			return modelsub.Subscription{}, nil
		},
		deleteFn: func(context.Context, uuid.UUID) error { return nil },
		listFn:   func(context.Context, modelsub.ListFilter) ([]modelsub.Subscription, int, error) { return nil, 0, nil },
		sumFn:    func(context.Context, modelsub.SummaryFilter) (int64, error) { return 0, nil },
	}

	log := logmid.NewLogger("error")
	h := New(log, u)
	r := Router(log, h)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/", bytes.NewBufferString("{not-json"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSummary_RequiresDates(t *testing.T) {
	u := &mockUsecase{
		sumFn: func(context.Context, modelsub.SummaryFilter) (int64, error) {
			t.Fatalf("usecase must not be called when from/to missing")
			return 0, nil
		},
		createFn: func(context.Context, modelsub.Subscription) (modelsub.Subscription, error) {
			return modelsub.Subscription{}, nil
		},
		getFn: func(context.Context, uuid.UUID) (modelsub.Subscription, error) { return modelsub.Subscription{}, nil },
		updateFn: func(context.Context, uuid.UUID, modelsub.Subscription) (modelsub.Subscription, error) {
			return modelsub.Subscription{}, nil
		},
		deleteFn: func(context.Context, uuid.UUID) error { return nil },
		listFn:   func(context.Context, modelsub.ListFilter) ([]modelsub.Subscription, int, error) { return nil, 0, nil },
	}

	log := logmid.NewLogger("error")
	h := New(log, u)
	r := Router(log, h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/summary", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want %d, got %d", http.StatusBadRequest, w.Code)
	}
}
