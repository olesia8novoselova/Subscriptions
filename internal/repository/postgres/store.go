package postgres

import (
	"context"
	"log/slog"

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
