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
	errValid   = errors.New("validation error")
	ErrOverlap = errors.New("overlapping subscription")
)

type SubscriptionRepository interface {
	Create(ctx context.Context, s *models.Subscription) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	List(ctx context.Context, f models.ListFilters) ([]models.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, id uuid.UUID, fields map[string]any) (*models.Subscription, error)
	FindActiveInPeriod(ctx context.Context, from, to time.Time, f models.ListFilters) ([]models.Subscription, error)
	ExistsOverlap(ctx context.Context, userID uuid.UUID, serviceName string, start time.Time, end *time.Time, excludeID *uuid.UUID) (bool, error)
}

type SubscriptionService struct {
	repo SubscriptionRepository
	log  *slog.Logger
}

func NewSubscriptionService(repo SubscriptionRepository, log *slog.Logger) *SubscriptionService {
	return &SubscriptionService{repo: repo, log: log}
}

// Create — создает новую подписку
// Проверяет пересечения с существующими подписками пользователя
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
		ID:          uuid.New(),
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   start,
		EndDate:     endPtr,
	}

	overlap, err := s.repo.ExistsOverlap(ctx, sub.UserID, sub.ServiceName, sub.StartDate, sub.EndDate, nil)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, ErrOverlap
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

// GetByID — получает подписку по ID
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

// List — получает список подписок с фильтрами и пагинацией
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
		UserID:      userIDPtr,
		ServiceName: serviceName,
		Limit:       limit,
		Offset:      offset,
	}
	return s.repo.List(ctx, f)
}

// Delete — удаляет подписку по ID
func (s *SubscriptionService) Delete(ctx context.Context, idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fmt.Errorf("%w: id must be UUID", errValid)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gorm.ErrRecordNotFound
		}
		return fmt.Errorf("db error: %w", err)
	}
	return nil
}

// Patch — обновляет подписку по ID
// Проверяет пересечения с существующими подписками пользователя
// Если поле пустое — не обновляет его
func (s *SubscriptionService) Patch(ctx context.Context, idStr string, req models.UpdateSubscriptionRequest) (*models.Subscription, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("%w: id must be UUID", errValid)
	}

	fields := make(map[string]any)

	if req.ServiceName != nil {
		if *req.ServiceName == "" {
			return nil, fmt.Errorf("%w: service_name cannot be empty", errValid)
		}
		fields["service_name"] = *req.ServiceName
	}

	if req.Price != nil {
		if *req.Price <= 0 {
			return nil, fmt.Errorf("%w: price must be positive integer", errValid)
		}
		fields["price"] = *req.Price
	}

	if req.StartDate != nil {
		start, err := parseMonthYear(*req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("%w: start_date must be MM-YYYY", errValid)
		}
		fields["start_date"] = start
	}

	if req.EndDate != nil {
		if *req.EndDate == "" {
			// очистить end_date
			fields["end_date"] = nil
		} else {
			end, err := parseMonthYear(*req.EndDate)
			if err != nil {
				return nil, fmt.Errorf("%w: end_date must be MM-YYYY or empty to clear", errValid)
			}
			// если обновляем обе даты — проверим порядок; если только end, проверим через БД
			if sd, ok := fields["start_date"].(time.Time); ok && end.Before(sd) {
				return nil, fmt.Errorf("%w: end_date must not be before start_date", errValid)
			}
			fields["end_date"] = end
		}
	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("%w: no fields to update", errValid)
	}

	// если задают только end_date — убедимся, что он не раньше текущего start_date
	if req.EndDate != nil && *req.EndDate != "" && fields["start_date"] == nil {
		existing, err := s.repo.FindByID(ctx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, gorm.ErrRecordNotFound
			}
			return nil, err
		}
		newEnd := fields["end_date"].(time.Time)
		if newEnd.Before(existing.StartDate) {
			return nil, fmt.Errorf("%w: end_date must not be before start_date", errValid)
		}
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	newStart := existing.StartDate
	if v, ok := fields["start_date"].(time.Time); ok {
		newStart = v
	}
	var newEnd *time.Time
	if _, ok := fields["end_date"]; ok {
		if fields["end_date"] == nil {
			newEnd = nil
		} else {
			v := fields["end_date"].(time.Time)
			newEnd = &v
		}
	} else {
		newEnd = existing.EndDate
	}

	overlap, err := s.repo.ExistsOverlap(ctx, existing.UserID, existing.ServiceName, newStart, newEnd, &id)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, ErrOverlap
	}

	return s.repo.Update(ctx, id, fields)
}

// TotalCost — суммарная стоимость за период [fromStr; toStr] c фильтрами
func (s *SubscriptionService) TotalCost(ctx context.Context, fromStr, toStr, userIDStr, serviceName string) (int, error) {
	from, err := parseMonthYear(fromStr) // "01-2006"
	if err != nil {
		return 0, fmt.Errorf("%w: from must be MM-YYYY", errValid)
	}
	to, err := parseMonthYear(toStr)
	if err != nil {
		return 0, fmt.Errorf("%w: to must be MM-YYYY", errValid)
	}
	if to.Before(from) {
		return 0, fmt.Errorf("%w: to must be >= from", errValid)
	}

	var userIDPtr *uuid.UUID
	if userIDStr != "" {
		uid, err := uuid.Parse(userIDStr)
		if err != nil {
			return 0, fmt.Errorf("%w: user_id must be UUID", errValid)
		}
		userIDPtr = &uid
	}

	f := models.ListFilters{
		UserID:      userIDPtr,
		ServiceName: serviceName,
		Limit:       0,
		Offset:      0,
	}

	subs, err := s.repo.FindActiveInPeriod(ctx, from, to, f)
	if err != nil {
		return 0, fmt.Errorf("db error: %w", err)
	}

	total := 0
	for _, sub := range subs {
		// нормализуем границы пересечения
		overlapStart := maxDate(sub.StartDate, from)
		overlapEnd := to
		if sub.EndDate != nil && sub.EndDate.Before(overlapEnd) {
			overlapEnd = *sub.EndDate
		}
		if overlapEnd.Before(overlapStart) {
			continue
		}
		months := monthsInclusive(overlapStart, overlapEnd)
		total += months * sub.Price
	}

	return total, nil
}

// monthsInclusive — количество месяцев между датами
func monthsInclusive(a, b time.Time) int {
	ay, am, _ := a.Date()
	by, bm, _ := b.Date()
	return (by-int(ay))*12 + int(bm-am) + 1
}

func maxDate(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
