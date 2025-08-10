package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/olesia8novoselova/Subscriptions/internal/models"
	"github.com/olesia8novoselova/Subscriptions/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) Create(ctx context.Context, s *models.Subscription) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *mockRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	args := m.Called(ctx, id)
	sub, _ := args.Get(0).(*models.Subscription)
	return sub, args.Error(1)
}

func (m *mockRepo) List(ctx context.Context, f models.ListFilters) ([]models.Subscription, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]models.Subscription), args.Error(1)
}

func (m *mockRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockRepo) Update(ctx context.Context, id uuid.UUID, fields map[string]any) (*models.Subscription, error) {
	args := m.Called(ctx, id, fields)
	sub, _ := args.Get(0).(*models.Subscription)
	return sub, args.Error(1)
}

func (m *mockRepo) FindActiveInPeriod(ctx context.Context, from, to time.Time, f models.ListFilters) ([]models.Subscription, error) {
	args := m.Called(ctx, from, to, f)
	return args.Get(0).([]models.Subscription), args.Error(1)
}

func (m *mockRepo) ExistsOverlap(ctx context.Context, userID uuid.UUID, serviceName string, start time.Time, end *time.Time, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, serviceName, start, end, excludeID)
	return args.Bool(0), args.Error(1)
}

// TestCreate_Valid - тестирует корректное создание подписки
func TestCreate_Valid(t *testing.T) {
	repo := new(mockRepo)
	svc := service.NewSubscriptionService(repo, nil)

	userID := uuid.New()
	start, _ := time.Parse("01-2006", "07-2025")
	req := models.CreateSubscriptionRequest{
		ServiceName: "Netflix",
		Price:       500,
		UserID:      userID.String(),
		StartDate:   "07-2025",
	}

	repo.On("ExistsOverlap", mock.Anything, userID, "Netflix", start, (*time.Time)(nil), (*uuid.UUID)(nil)).Return(false, nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*models.Subscription")).Return(nil)

	sub, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, "Netflix", sub.ServiceName)
	repo.AssertExpectations(t)
}

// TestCreate_InvalidDate - тестирует создание подписки с некорректной датой
func TestCreate_InvalidDate(t *testing.T) {
	repo := new(mockRepo)
	svc := service.NewSubscriptionService(repo, nil)

	req := models.CreateSubscriptionRequest{
		ServiceName: "Netflix",
		Price:       500,
		UserID:      uuid.New().String(),
		StartDate:   "invalid-date",
	}

	sub, err := svc.Create(context.Background(), req)
	assert.Nil(t, sub)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation error")
}

// TestCreate_Overlap - тестирует создание подписки с пересечением
func TestCreate_Overlap(t *testing.T) {
	repo := new(mockRepo)
	svc := service.NewSubscriptionService(repo, nil)

	userID := uuid.New()
	start, _ := time.Parse("01-2006", "07-2025")
	req := models.CreateSubscriptionRequest{
		ServiceName: "Netflix",
		Price:       500,
		UserID:      userID.String(),
		StartDate:   "07-2025",
	}

	repo.On("ExistsOverlap", mock.Anything, userID, "Netflix", start, (*time.Time)(nil), (*uuid.UUID)(nil)).Return(true, nil)

	sub, err := svc.Create(context.Background(), req)
	assert.Nil(t, sub)
	assert.ErrorIs(t, err, service.ErrOverlap)
}

// TestGetByID_Success - тестирует получение подписки по ID
func TestGetByID_NotFound(t *testing.T) {
	repo := new(mockRepo)
	svc := service.NewSubscriptionService(repo, nil)

	id := uuid.New()
	repo.On("FindByID", mock.Anything, id).Return(nil, gorm.ErrRecordNotFound)

	sub, err := svc.GetByID(context.Background(), id.String())
	assert.Nil(t, sub)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// TestList_DefaultLimitOffset - тестирует получение списка подписок со значениями по умолчанию для limit и offset
func TestList_DefaultLimitOffset(t *testing.T) {
	repo := new(mockRepo)
	svc := service.NewSubscriptionService(repo, nil)

	expected := []models.Subscription{{ServiceName: "Test"}}
	repo.On("List", mock.Anything, mock.Anything).Return(expected, nil)

	list, err := svc.List(context.Background(), "", "", 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, expected, list)
}

// TestDelete_Success - тестирует успешное удаление подписки
func TestDelete_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := service.NewSubscriptionService(repo, nil)

	id := uuid.New()
	repo.On("Delete", mock.Anything, id).Return(nil)

	err := svc.Delete(context.Background(), id.String())
	assert.NoError(t, err)
}

// TestPatch_Overlap - тестирует обновление подписки с пересечением
func TestPatch_Overlap(t *testing.T) {
	repo := new(mockRepo)
	svc := service.NewSubscriptionService(repo, nil)

	id := uuid.New()
	start := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	existing := &models.Subscription{
		ID:          id,
		ServiceName: "Test",
		UserID:      uuid.New(),
		StartDate:   start,
	}

	repo.On("FindByID", mock.Anything, id).Return(existing, nil).Once()
	repo.On("FindByID", mock.Anything, id).Return(existing, nil).Once()
	repo.On("ExistsOverlap", mock.Anything, existing.UserID, existing.ServiceName, start, (*time.Time)(nil), &id).Return(true, nil)

	req := models.UpdateSubscriptionRequest{ServiceName: strPtr("NewName")}
	sub, err := svc.Patch(context.Background(), id.String(), req)

	assert.Nil(t, sub)
	assert.ErrorIs(t, err, service.ErrOverlap)
}

// TestTotalCost_Calculation - тестирует корректный расчет суммарной стоимости подписок
func TestTotalCost_Calculation(t *testing.T) {
	repo := new(mockRepo)
	svc := service.NewSubscriptionService(repo, nil)

	from, _ := time.Parse("01-2006", "07-2025")
	to, _ := time.Parse("01-2006", "09-2025")

	subs := []models.Subscription{
		{Price: 100, StartDate: from, EndDate: ptrTime(to)}, // 3 месяца
	}
	repo.On("FindActiveInPeriod", mock.Anything, from, to, mock.Anything).Return(subs, nil)

	total, err := svc.TotalCost(context.Background(), "07-2025", "09-2025", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 300, total)
}

func strPtr(s string) *string        { return &s }
func ptrTime(t time.Time) *time.Time { return &t }
