package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/olesia8novoselova/Subscriptions/internal/models"
	"gorm.io/gorm"
)

var (
	errValid = errors.New("validation error")
)

type SubscriptionRepository interface {
	Create(ctx context.Context, s *models.Subscription) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	List(ctx context.Context, f models.ListFilters) ([]models.Subscription, error)
}

type SubscriptionService struct {
	repo SubscriptionRepository
	log  *slog.Logger
}

func NewSubscriptionService(repo SubscriptionRepository, log *slog.Logger) *SubscriptionService {
	return &SubscriptionService{repo: repo, log: log}
}

func (s *SubscriptionService) Create(ctx context.Context, req models.CreateSubscriptionRequest) (*models.Subscription, error) {
	// Валидация
	if req.ServiceName == "" {
		return nil, fmt.Errorf("%w: service_name is required", errValid)
	}
	if req.Price <= 0 {
		return nil, fmt.Errorf("%w: price must be positive integer", errValid)
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: user_id must be UUID", errValid)
	}

	start, err := parseMonthYear(req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("%w: start_date format must be MM-YYYY or YYYY-MM", errValid)
	}

	var endPtr *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		end, err := parseMonthYear(*req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("%w: end_date format must be MM-YYYY or YYYY-MM", errValid)
		}
		if end.Before(start) {
			return nil, fmt.Errorf("%w: end_date must not be before start_date", errValid)
		}
		endPtr = &end
	}

	sub := &models.Subscription{
		ID: uuid.New(),
		ServiceName: req.ServiceName,
		Price: req.Price,
		UserID: userID,
		StartDate: start,
		EndDate: endPtr,
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

// Принимает "MM-YYYY" или "YYYY-MM", возвращает 1-е число месяца в UTC
func parseMonthYear(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	if t, err := time.Parse("01-2006", s); err == nil {
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
	}
	if t, err := time.Parse("2006-01", s); err == nil {
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
	}
	return time.Time{}, fmt.Errorf("invalid format")
}

func (s *SubscriptionService) GetByID(ctx context.Context, idStr string) (*models.Subscription, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("%w: id must be UUID", errValid)
	}
	sub, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("db error: %w", err)
	}
	return sub, nil
}

func (s *SubscriptionService) List(ctx context.Context, userIDStr, serviceName string, limit, offset int) ([]models.Subscription, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	var userIDPtr *uuid.UUID
	if userIDStr != "" {
		uid, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, fmt.Errorf("%w: user_id must be UUID", errValid)
		}
		userIDPtr = &uid
	}

	f := models.ListFilters{
		UserID: userIDPtr,
		ServiceName: serviceName,
		Limit: limit,
		Offset: offset,
	}
	return s.repo.List(ctx, f)
}