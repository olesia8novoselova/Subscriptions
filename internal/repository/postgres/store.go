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