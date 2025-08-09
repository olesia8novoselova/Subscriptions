package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/olesia8novoselova/Subscriptions/internal/models"
)

var (
	errValid = errors.New("validation error")
)

type SubscriptionRepository interface {
	Create(ctx context.Context, s *models.Subscription) error
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
