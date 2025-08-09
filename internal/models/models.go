package models

import (
	"time"

	"github.com/google/uuid"
)

// Subscription — основная модель подписки в БД
type Subscription struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	ServiceName string `json:"service_name" gorm:"type:text;not null"`
	Price int `json:"price" gorm:"type:int;not null"`
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	StartDate time.Time `json:"start_date" gorm:"type:date;not null"`
	EndDate *time.Time `json:"end_date,omitempty" gorm:"type:date"`

	CreatedAt time.Time `json:"-" gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt time.Time `json:"-" gorm:"type:timestamptz;not null;default:now()"`
}

// CreateSubscriptionRequest — тело запроса на создание подписки
type CreateSubscriptionRequest struct {
	ServiceName string `json:"service_name"`
	Price int `json:"price"`
	UserID string `json:"user_id"`
	StartDate string `json:"start_date"`
	EndDate *string `json:"end_date,omitempty"`
}

// SubscriptionResponse — ответ на запрос подписки
type SubscriptionResponse struct {
	ID uuid.UUID `json:"id"`
	ServiceName string `json:"service_name"`
	Price int `json:"price"`
	UserID  uuid.UUID `json:"user_id"`
	StartDate  string `json:"start_date"`
	EndDate  *string `json:"end_date,omitempty"`
}

// ListFilters — фильтры для списка подписок
// Используется для пагинации и фильтрации по полям
type ListFilters struct {
	UserID *uuid.UUID
	ServiceName string
	Limit int
	Offset int
}