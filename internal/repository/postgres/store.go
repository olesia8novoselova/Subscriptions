package postgres

import (
	"context"
	"errors"
	"log/slog"

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