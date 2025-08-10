package controller

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/olesia8novoselova/Subscriptions/internal/models"
	"github.com/olesia8novoselova/Subscriptions/internal/service"
	"gorm.io/gorm"
)

type SubscriptionHandler struct {
	svc *service.SubscriptionService
	log *slog.Logger
}

func NewSubscriptionHandler(svc *service.SubscriptionService, log *slog.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc, log: log}
}

// CreateSubscription
// @Summary Create subscription
// @Description Создаёт запись о подписке
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param  request  body models.CreateSubscriptionRequest  true  "Subscription body"
// @Success  201  {object}  models.SubscriptionResponse
// @Failure  400  {object}  map[string]string
// @Router  /api/subscriptions  [post]
func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	sub, err := h.svc.Create(r.Context(), req)
	if err != nil {
		h.log.Error("create subscription failed", "error", err)
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := toResponse(sub)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}


// GetSubscription
// @Summary Get subscription by id
// @Description  Возвращает запись по её ID
// @Tags subscriptions
// @Produce json
// @Param  id  path  string  true  "Subscription ID (UUID)"  example("b548150d-6198-4cc1-a186-8c4a1e0ccdcf")
// @Success  200 {object}  models.SubscriptionResponse
// @Failure  400 {object}  map[string]string
// @Failure  404 {object}  map[string]string
// @Router  /api/subscriptions/{id}  [get]
func (h *SubscriptionHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Получаем id из пути
	id := strings.TrimPrefix(r.URL.Path, "/api/subscriptions/")
	if id == "" || strings.Contains(id, "/") {
		h.writeError(w, http.StatusBadRequest, "invalid id path")
		return
	}

	sub, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			h.writeError(w, http.StatusNotFound, "subscription not found")
			return
		}
		h.log.Error("get subscription failed", "error", err)
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := toResponse(sub)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}


// ListSubscriptions
// @Summary List subscriptions
// @Description Список подписок с фильтрами и пагинацией
// @Tags subscriptions
// @Produce json
// @Param  user_id  query  string  false  "Filter by user UUID"  example("b548150d-6198-4cc1-a186-8c4a1e0ccdcf")
// @Param  service_name  query  string false  "Filter by service name"  example("Test Service")  default("Test Service")
// @Param  limit  query  int  false  "Page size (default 20, max 100)"  example(20)  default(20)
// @Param  offset  query  int  false  "Offset (default 0)" example(0)  default(0)
// @Success  200 {array}  models.SubscriptionResponse
// @Failure  400 {object}  map[string]string
// @Router  /api/subscriptions  [get]
func (h *SubscriptionHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()
	userID := q.Get("user_id")
	serviceName := q.Get("service_name")

	limit := 0
	offset := 0
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		} else {
			h.writeError(w, http.StatusBadRequest, "limit must be integer")
			return
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			offset = n
		} else {
			h.writeError(w, http.StatusBadRequest, "offset must be integer")
			return
		}
	}

	list, err := h.svc.List(r.Context(), userID, serviceName, limit, offset)
	if err != nil {
		h.log.Error("list subscriptions failed", "error", err)
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := make([]models.SubscriptionResponse, 0, len(list))
	for _, s := range list {
		resp = append(resp, toResponse(&s))
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}


// DeleteSubscription
// @Summary Delete subscription by id
// @Description Удаляет подписку по её ID
// @Tags subscriptions
// @Param  id  path  string  true "Subscription ID (UUID)"  example("b548150d-6198-4cc1-a186-8c4a1e0ccdcf")
// @Success  204  "No Content"
// @Failure  400  {object}  map[string]string
// @Failure  404  {object}  map[string]string
// @Router  /api/subscriptions/{id}  [delete]
func (h *SubscriptionHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/subscriptions/")
	if id == "" || strings.Contains(id, "/") {
		h.writeError(w, http.StatusBadRequest, "invalid id path")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		if err == gorm.ErrRecordNotFound {
			h.writeError(w, http.StatusNotFound, "subscription not found")
			return
		}
		h.log.Error("delete subscription failed", "error", err)
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PatchSubscription
// @Summary Patch subscription
// @Description Частичное обновление полей подписки. Чтобы очистить end_date, передайте "".
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id  path  string  true  "Subscription ID"  format(uuid)  example("b548150d-6198-4cc1-a186-8c4a1e0ccdcf")
// @Param request body  models.UpdateSubscriptionRequest  true  "Fields to update"
// @Success  200  {object}  models.SubscriptionResponse
// @Failure  400  {object}  map[string]string
// @Failure  404  {object}  map[string]string
// @Router /api/subscriptions/{id}  [patch]
func (h *SubscriptionHandler) PatchSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/subscriptions/")
	if id == "" || strings.Contains(id, "/") {
		h.writeError(w, http.StatusBadRequest, "invalid id path")
		return
	}

	var req models.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	sub, err := h.svc.Patch(r.Context(), id, req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			h.writeError(w, http.StatusNotFound, "subscription not found")
			return
		}
		h.log.Error("patch subscription failed", "error", err)
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := toResponse(sub)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}



func (h *SubscriptionHandler) writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func toResponse(s *models.Subscription) models.SubscriptionResponse {
	start := s.StartDate.Format("01-2006")
	var endStr *string
	if s.EndDate != nil {
		es := s.EndDate.Format("01-2006")
		endStr = &es
	}
	return models.SubscriptionResponse{
		ID: s.ID,
		ServiceName: s.ServiceName,
		Price: s.Price,
		UserID:  s.UserID,
		StartDate: start,
		EndDate: endStr,
	}
}
