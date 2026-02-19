package subscription

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	modeldate "test_task/internal/domain/models/month_year"
	modelsub "test_task/internal/domain/models/subscription"
	JSONRes "test_task/pkg/JSON_response"
	myerrors "test_task/pkg/global_errors"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UsecaseI interface {
	Create(ctx context.Context, s modelsub.Subscription) (modelsub.Subscription, error)
	GetSub(ctx context.Context, id uuid.UUID) (modelsub.Subscription, error)
	UpdateSub(ctx context.Context, id uuid.UUID, s modelsub.Subscription) (modelsub.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, f modelsub.ListFilter) ([]modelsub.Subscription, int, error)
	Summary(ctx context.Context, f modelsub.SummaryFilter) (int64, error)
}

type Handler struct {
	log     *slog.Logger
	usecase UsecaseI
}

func New(log *slog.Logger, usecase UsecaseI) *Handler {
	return &Handler{log: log, usecase: usecase}
}

func toResp(s modelsub.Subscription) modelsub.SubscriptionResp {
	var end *string
	if s.EndDate != nil {
		v := modeldate.FormatMonthYear(*s.EndDate)
		end = &v
	}
	return modelsub.SubscriptionResp{
		ID:          s.ID.String(),
		ServiceName: s.ServiceName,
		Price:       s.Price,
		UserID:      s.UserID.String(),
		StartDate:   modeldate.FormatMonthYear(s.StartDate),
		EndDate:     end,
		CreatedAt:   s.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   s.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	JSONRes.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req modelsub.SubscriptionCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONRes.WriteJSON(w, http.StatusBadRequest, "invalid json body")
		return
	}

	req.ServiceName = strings.TrimSpace(req.ServiceName)
	if req.ServiceName == "" {
		JSONRes.WriteJSON(w, http.StatusBadRequest, "service_name is required")
		return
	}
	if req.Price < 0 {
		JSONRes.WriteJSON(w, http.StatusBadRequest, "price must be >= 0")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		JSONRes.WriteJSON(w, http.StatusBadRequest, "user_id must be a valid UUID")
		return
	}

	start, err := modeldate.ParseMonthYear(req.StartDate)
	if err != nil {
		JSONRes.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	var end *time.Time
	if req.EndDate != nil {
		t, err := modeldate.ParseMonthYear(*req.EndDate)
		if err != nil {
			JSONRes.WriteJSON(w, http.StatusBadRequest, err.Error())
			return
		}
		if t.Before(start) {
			JSONRes.WriteJSON(w, http.StatusBadRequest, "end_date must be >= start_date")
			return
		}
		end = &t
	}

	s := modelsub.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   start,
		EndDate:     end,
	}

	created, err := h.usecase.Create(r.Context(), s)
	if err != nil {
		h.log.Error("create subscription failed", slog.Any("err", err))
		JSONRes.WriteJSON(w, http.StatusConflict, "failed to create subscription")
		return
	}

	JSONRes.WriteJSON(w, http.StatusCreated, toResp(created))
}

func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		JSONRes.WriteJSON(w, http.StatusBadRequest, "id must be a valid UUID")
		return
	}

	s, err := h.usecase.GetSub(r.Context(), id)
	if err != nil {
		if errors.Is(err, myerrors.ErrorNotFound) {
			JSONRes.WriteJSON(w, http.StatusNotFound, "subscription not found")
			return
		}
		h.log.Error("get subscription failed", slog.Any("err", err))
		JSONRes.WriteJSON(w, http.StatusPreconditionFailed, "failed to get subscription")
		return
	}

	JSONRes.WriteJSON(w, http.StatusOK, toResp(s))
}

func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		JSONRes.WriteJSON(w, http.StatusBadRequest, "id must be a valid UUID")
		return
	}

	var req modelsub.SubscriptionCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONRes.WriteJSON(w, http.StatusBadRequest, "invalid json body")
		return
	}

	req.ServiceName = strings.TrimSpace(req.ServiceName)
	if req.ServiceName == "" {
		JSONRes.WriteJSON(w, http.StatusBadRequest, "service_name is required")
		return
	}
	if req.Price < 0 {
		JSONRes.WriteJSON(w, http.StatusBadRequest, "price must be >= 0")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		JSONRes.WriteJSON(w, http.StatusBadRequest, "user_id must be a valid UUID")
		return
	}

	start, err := modeldate.ParseMonthYear(req.StartDate)
	if err != nil {
		JSONRes.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	var end *time.Time
	if req.EndDate != nil {
		t, err := modeldate.ParseMonthYear(*req.EndDate)
		if err != nil {
			JSONRes.WriteJSON(w, http.StatusBadRequest, err.Error())
			return
		}
		if t.Before(start) {
			JSONRes.WriteJSON(w, http.StatusBadRequest, "end_date must be >= start_date")
			return
		}
		end = &t
	}

	s := modelsub.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   start,
		EndDate:     end,
	}

	updated, err := h.usecase.UpdateSub(r.Context(), id, s)
	if err != nil {
		if errors.Is(err, myerrors.ErrorNotFound) {
			JSONRes.WriteJSON(w, http.StatusNotFound, "subscription not found")
			return
		}
		h.log.Error("update subscription failed", slog.Any("err", err))
		JSONRes.WriteJSON(w, http.StatusForbidden, "failed to update subscription")
		return
	}

	JSONRes.WriteJSON(w, http.StatusOK, toResp(updated))
}

func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		JSONRes.WriteJSON(w, http.StatusBadRequest, "id must be a valid UUID")
		return
	}

	if err := h.usecase.Delete(r.Context(), id); err != nil {
		if errors.Is(err, myerrors.ErrorNotFound) {
			JSONRes.WriteJSON(w, http.StatusNotFound, "subscription not found")
			return
		}
		h.log.Error("delete subscription failed", slog.Any("err", err))
		JSONRes.WriteJSON(w, http.StatusForbidden, "failed to delete subscription")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var userID *uuid.UUID
	if v := strings.TrimSpace(q.Get("user_id")); v != "" {
		parsed, err := uuid.Parse(v)
		if err != nil {
			JSONRes.WriteJSON(w, http.StatusBadRequest, "user_id must be a valid UUID")
			return
		}
		userID = &parsed
	}

	var serviceName *string
	if v := strings.TrimSpace(q.Get("service_name")); v != "" {
		serviceName = &v
	}

	limit := 50
	offset := 0
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			offset = n
		}
	}

	items, total, err := h.usecase.List(r.Context(), modelsub.ListFilter{
		UserID:      userID,
		ServiceName: serviceName,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		h.log.Error("list subscriptions failed", slog.Any("err", err))
		JSONRes.WriteJSON(w, http.StatusForbidden, "failed to list subscriptions")
		return
	}

	respItems := make([]modelsub.SubscriptionResp, 0, len(items))
	for _, s := range items {
		respItems = append(respItems, toResp(s))
	}

	JSONRes.WriteJSON(w, http.StatusOK, map[string]any{
		"total":  total,
		"limit":  limit,
		"offset": offset,
		"items":  respItems,
	})
}

func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	fromStr := strings.TrimSpace(q.Get("from"))
	toStr := strings.TrimSpace(q.Get("to"))
	if fromStr == "" || toStr == "" {
		JSONRes.WriteJSON(w, http.StatusBadRequest, "from and to are required (MM-YYYY)")
		return
	}

	from, err := modeldate.ParseMonthYear(fromStr)
	if err != nil {
		JSONRes.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	to, err := modeldate.ParseMonthYear(toStr)
	if err != nil {
		JSONRes.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	if to.Before(from) {
		JSONRes.WriteJSON(w, http.StatusBadRequest, "to must be >= from")
		return
	}

	var userID *uuid.UUID
	if v := strings.TrimSpace(q.Get("user_id")); v != "" {
		parsed, err := uuid.Parse(v)
		if err != nil {
			JSONRes.WriteJSON(w, http.StatusBadRequest, "user_id must be a valid UUID")
			return
		}
		userID = &parsed
	}

	var serviceName *string
	if v := strings.TrimSpace(q.Get("service_name")); v != "" {
		serviceName = &v
	}

	total, err := h.usecase.Summary(r.Context(), modelsub.SummaryFilter{
		From:        from,
		To:          to,
		UserID:      userID,
		ServiceName: serviceName,
	})
	if err != nil {
		h.log.Error("summary failed", slog.Any("err", err))
		JSONRes.WriteJSON(w, http.StatusForbidden, "failed to calculate summary")
		return
	}

	resp := map[string]any{
		"total":    total,
		"currency": "RUB",
		"from":     modeldate.FormatMonthYear(from),
		"to":       modeldate.FormatMonthYear(to),
	}
	if userID != nil {
		resp["user_id"] = userID.String()
	}
	if serviceName != nil {
		resp["service_name"] = *serviceName
	}

	JSONRes.WriteJSON(w, http.StatusOK, resp)
}
