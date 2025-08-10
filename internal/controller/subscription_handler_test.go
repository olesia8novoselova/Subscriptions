package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olesia8novoselova/Subscriptions/internal/controller"
	"github.com/olesia8novoselova/Subscriptions/internal/models"
	"github.com/olesia8novoselova/Subscriptions/internal/service"
	"gorm.io/gorm"
)

type fakeService struct {
	CreateFn    func(ctx context.Context, req models.CreateSubscriptionRequest) (*models.Subscription, error)
	GetByIDFn   func(ctx context.Context, id string) (*models.Subscription, error)
	ListFn      func(ctx context.Context, userID, serviceName string, limit, offset int) ([]models.Subscription, error)
	DeleteFn    func(ctx context.Context, id string) error
	PatchFn     func(ctx context.Context, id string, req models.UpdateSubscriptionRequest) (*models.Subscription, error)
	TotalCostFn func(ctx context.Context, from, to, userID, serviceName string) (int, error)
}

func (f *fakeService) Create(ctx context.Context, req models.CreateSubscriptionRequest) (*models.Subscription, error) {
	return f.CreateFn(ctx, req)
}
func (f *fakeService) GetByID(ctx context.Context, id string) (*models.Subscription, error) {
	return f.GetByIDFn(ctx, id)
}
func (f *fakeService) List(ctx context.Context, userID, serviceName string, limit, offset int) ([]models.Subscription, error) {
	return f.ListFn(ctx, userID, serviceName, limit, offset)
}
func (f *fakeService) Delete(ctx context.Context, id string) error {
	return f.DeleteFn(ctx, id)
}
func (f *fakeService) Patch(ctx context.Context, id string, req models.UpdateSubscriptionRequest) (*models.Subscription, error) {
	return f.PatchFn(ctx, id, req)
}
func (f *fakeService) TotalCost(ctx context.Context, from, to, userID, serviceName string) (int, error) {
	return f.TotalCostFn(ctx, from, to, userID, serviceName)
}

func mustUUID(s string) uuid.UUID { u, _ := uuid.Parse(s); return u }

func subDTO() *models.Subscription {
	id := mustUUID("b548150d-6198-4cc1-a186-8c4a1e0ccdcf")
	user := mustUUID("60601fee-2bf1-4721-ae6f-7636e79a0cba")
	start := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	return &models.Subscription{
		ID:          id,
		ServiceName: "Test Service",
		Price:       500,
		UserID:      user,
		StartDate:   start,
	}
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// TestCreateSubscription_Success - тестирует успешное создание подписки
func TestCreateSubscription_Success(t *testing.T) {
	fs := &fakeService{
		CreateFn: func(ctx context.Context, req models.CreateSubscriptionRequest) (*models.Subscription, error) {
			return subDTO(), nil
		},
	}
	h := controller.NewSubscriptionHandler(fs, newTestLogger())

	body := `{"service_name":"Test Service","price":500,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"07-2025"}`
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.CreateSubscription(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", res.StatusCode)
	}
	var got models.SubscriptionResponse
	_ = json.NewDecoder(res.Body).Decode(&got)
	if got.ServiceName != "Test Service" || got.Price != 500 {
		t.Fatalf("unexpected body: %+v", got)
	}
}

// TestCreateSubscription_Conflict - тестирует конфликт при создании подписки
func TestCreateSubscription_Conflict(t *testing.T) {
	fs := &fakeService{
		CreateFn: func(ctx context.Context, req models.CreateSubscriptionRequest) (*models.Subscription, error) {
			return nil, service.ErrOverlap
		},
	}
	h := controller.NewSubscriptionHandler(fs, newTestLogger())

	body := `{"service_name":"Test Service","price":500,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"07-2025"}`
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.CreateSubscription(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", w.Code)
	}
}

// TestGetSubscription_NotFound - тестирует получение подписки по ID, когда она не найдена
func TestGetSubscription_NotFound(t *testing.T) {
	fs := &fakeService{
		GetByIDFn: func(ctx context.Context, id string) (*models.Subscription, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	h := controller.NewSubscriptionHandler(fs, newTestLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/subscriptions/b548150d-6198-4cc1-a186-8c4a1e0ccdcf", nil)
	w := httptest.NewRecorder()

	h.GetSubscription(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

// TestGetSubscription_OK - тестирует успешное получение подписки по ID
func TestGetSubscription_OK(t *testing.T) {
	fs := &fakeService{
		GetByIDFn: func(ctx context.Context, id string) (*models.Subscription, error) {
			return subDTO(), nil
		},
	}
	h := controller.NewSubscriptionHandler(fs, newTestLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/subscriptions/b548150d-6198-4cc1-a186-8c4a1e0ccdcf", nil)
	w := httptest.NewRecorder()

	h.GetSubscription(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

// TestListSubscriptions_OK - тестирует получение списка подписок
func TestListSubscriptions_OK(t *testing.T) {
	fs := &fakeService{
		ListFn: func(ctx context.Context, user, svc string, limit, offset int) ([]models.Subscription, error) {
			return []models.Subscription{*subDTO()}, nil
		},
	}
	h := controller.NewSubscriptionHandler(fs, newTestLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/subscriptions?limit=10&offset=0&service_name=Test", nil)
	w := httptest.NewRecorder()

	h.ListSubscriptions(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var arr []models.SubscriptionResponse
	_ = json.NewDecoder(w.Body).Decode(&arr)
	if len(arr) != 1 || arr[0].ServiceName != "Test Service" {
		t.Fatalf("unexpected body: %+v", arr)
	}
}

// TestDeleteSubscription_OK - тестирует успешное удаление подписки
func TestDeleteSubscription_OK(t *testing.T) {
	fs := &fakeService{
		DeleteFn: func(ctx context.Context, id string) error { return nil },
	}
	h := controller.NewSubscriptionHandler(fs, newTestLogger())

	req := httptest.NewRequest(http.MethodDelete, "/api/subscriptions/b548150d-6198-4cc1-a186-8c4a1e0ccdcf", nil)
	w := httptest.NewRecorder()

	h.DeleteSubscription(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", w.Code)
	}
}

// TestDeleteSubscription_NotFound - тестирует удаление подписки, когда она не найдена
func TestDeleteSubscription_NotFound(t *testing.T) {
	fs := &fakeService{
		DeleteFn: func(ctx context.Context, id string) error { return gorm.ErrRecordNotFound },
	}
	h := controller.NewSubscriptionHandler(fs, newTestLogger())

	req := httptest.NewRequest(http.MethodDelete, "/api/subscriptions/b548150d-6198-4cc1-a186-8c4a1e0ccdcf", nil)
	w := httptest.NewRecorder()

	h.DeleteSubscription(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

// TestPatchSubscription_Conflict - тестирует конфликт при обновлении подписки
func TestPatchSubscription_Conflict(t *testing.T) {
	fs := &fakeService{
		PatchFn: func(ctx context.Context, id string, req models.UpdateSubscriptionRequest) (*models.Subscription, error) {
			return nil, service.ErrOverlap
		},
	}
	h := controller.NewSubscriptionHandler(fs, newTestLogger())

	body := `{"price":600}`
	req := httptest.NewRequest(http.MethodPatch, "/api/subscriptions/b548150d-6198-4cc1-a186-8c4a1e0ccdcf", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.PatchSubscription(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", w.Code)
	}
}

// TestPatchSubscription_OK - тестирует успешное обновление подписки
func TestPatchSubscription_OK(t *testing.T) {
	fs := &fakeService{
		PatchFn: func(ctx context.Context, id string, req models.UpdateSubscriptionRequest) (*models.Subscription, error) {
			s := subDTO()
			s.Price = 600
			return s, nil
		},
	}
	h := controller.NewSubscriptionHandler(fs, newTestLogger())

	body := `{"price":600}`
	req := httptest.NewRequest(http.MethodPatch, "/api/subscriptions/b548150d-6198-4cc1-a186-8c4a1e0ccdcf", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.PatchSubscription(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

// TestGetTotalCost_OK - тестирует получение суммарной стоимости подписок за период
func TestGetTotalCost_OK(t *testing.T) {
	fs := &fakeService{
		TotalCostFn: func(ctx context.Context, from, to, user, svc string) (int, error) {
			return 1450, nil
		},
	}
	h := controller.NewSubscriptionHandler(fs, newTestLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/subscriptions/total?from=07-2025&to=09-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba", nil)
	w := httptest.NewRecorder()

	h.GetTotalCost(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var got map[string]int
	_ = json.NewDecoder(w.Body).Decode(&got)
	if got["total"] != 1450 {
		t.Fatalf("total = %d, want 1450", got["total"])
	}
}

// TestGetTotalCost_ValidationError - тестирует ошибку валидации при получении суммарной стоимости
func TestGetTotalCost_ValidationError(t *testing.T) {
	fs := &fakeService{
		TotalCostFn: func(ctx context.Context, from, to, user, svc string) (int, error) {
			return 0, errors.New("validation error")
		},
	}
	h := controller.NewSubscriptionHandler(fs, newTestLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/subscriptions/total?from=bad&to=bad", nil)
	w := httptest.NewRecorder()

	h.GetTotalCost(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}
