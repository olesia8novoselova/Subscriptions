package controller

import (
	"encoding/json"
	"log/slog"
	"net/http"
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
// @Param request body models.CreateSubscriptionRequest true "Subscription body"
// @Success 201 {object} models.SubscriptionResponse
// @Failure 400 {object} map[string]string
// @Router /api/subscriptions [post]
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
// @Param id path string true "Subscription ID (UUID)"
// @Success 200 {object} models.SubscriptionResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/subscriptions/{id} [get]
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
