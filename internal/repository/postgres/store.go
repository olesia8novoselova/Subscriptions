package postgres

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/olesia8novoselova/Subscriptions/internal/models"
	"gorm.io/gorm"
)

type SubscriptionRepo struct {
	db  *gorm.DB
	log *slog.Logger
}

func New(db *gorm.DB, log *slog.Logger) *SubscriptionRepo {
	return &SubscriptionRepo{db: db, log: log}
}

func (r *SubscriptionRepo) Create(ctx context.Context, s *models.Subscription) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *SubscriptionRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	var sub models.Subscription
	err := r.db.WithContext(ctx).First(&sub, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &sub, err
}

func (r *SubscriptionRepo) List(ctx context.Context, f models.ListFilters) ([]models.Subscription, error) {
	q := r.db.WithContext(ctx).Model(&models.Subscription{})

	if f.UserID != nil {
		q = q.Where("user_id = ?", *f.UserID)
	}
	if f.ServiceName != "" {
		q = q.Where("service_name ILIKE ?", "%"+f.ServiceName+"%")
	}

	var res []models.Subscription
	err := q.Order("start_date DESC, created_at DESC").
		Limit(f.Limit).Offset(f.Offset).
		Find(&res).Error
	return res, err
}

func (r *SubscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&models.Subscription{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *SubscriptionRepo) Update(ctx context.Context, id uuid.UUID, fields map[string]any) (*models.Subscription, error) {
	tx := r.db.WithContext(ctx).Model(&models.Subscription{}).Where("id = ?", id).Updates(fields)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var sub models.Subscription
	if err := r.db.WithContext(ctx).First(&sub, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &sub, nil
}

// FindActiveInPeriod — подписки, которые пересекают период [from, to].
func (r *SubscriptionRepo) FindActiveInPeriod(ctx context.Context, from, to time.Time, f models.ListFilters) ([]models.Subscription, error) {
	q := r.db.WithContext(ctx).Model(&models.Subscription{})

	if f.UserID != nil {
		q = q.Where("user_id = ?", *f.UserID)
	}
	if f.ServiceName != "" {
		q = q.Where("service_name ILIKE ?", "%"+f.ServiceName+"%")
	}

	// Пересечение интервалов
	q = q.Where("start_date <= ?", to).
		Where("(end_date IS NULL OR end_date >= ?)", from)

	var res []models.Subscription
	if err := q.Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

// ExistsOverlap — проверяет, есть ли пересечение по (user_id, service_name) с данным периодом
func (r *SubscriptionRepo) ExistsOverlap(ctx context.Context, userID uuid.UUID, serviceName string, start time.Time, end *time.Time, excludeID *uuid.UUID) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&models.Subscription{}).
		Where("user_id = ?", userID).
		Where("lower(service_name) = ?", strings.ToLower(serviceName)).
		Where("start_date <= ?", coalesceEnd(end)).
		Where("(end_date IS NULL OR end_date >= ?)", start)

	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}

	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func coalesceEnd(end *time.Time) time.Time {
	if end == nil {
		// далеко в будущем, чтобы условие start_date <= end выполнялось для всех
		return time.Date(9999, 12, 1, 0, 0, 0, 0, time.UTC)
	}
	return *end
}
