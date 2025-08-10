package models

import (
	"time"

	"github.com/google/uuid"
)

// Subscription — основная модель подписки в БД
type Subscription struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	ServiceName string     `json:"service_name" gorm:"type:text;not null"`
	Price       int        `json:"price" gorm:"type:int;not null"`
	UserID      uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	StartDate   time.Time  `json:"start_date" gorm:"type:date;not null"`
	EndDate     *time.Time `json:"end_date,omitempty" gorm:"type:date"`

	CreatedAt time.Time `json:"-" gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt time.Time `json:"-" gorm:"type:timestamptz;not null;default:now()"`
}

// CreateSubscriptionRequest — тело запроса на создание подписки
type CreateSubscriptionRequest struct {
	ServiceName string  `json:"service_name" example:"Test Service"`
	Price       int     `json:"price" example:"500"`
	UserID      string  `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string  `json:"start_date" example:"07-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"09-2025"`
}

// SubscriptionResponse — ответ на запрос подписки
type SubscriptionResponse struct {
	ID          uuid.UUID `json:"id" example:"b548150d-6198-4cc1-a186-8c4a1e0ccdcf"`
	ServiceName string    `json:"service_name" example:"Test Service"`
	Price       int       `json:"price" example:"500"`
	UserID      uuid.UUID `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string    `json:"start_date" example:"07-2025"`
	EndDate     *string   `json:"end_date,omitempty" example:"09-2025"`
}

// ListFilters — фильтры для списка подписок
// Используется для пагинации и фильтрации по полям
type ListFilters struct {
	UserID      *uuid.UUID
	ServiceName string
	Limit       int
	Offset      int
}

type UpdateSubscriptionRequest struct {
	ServiceName *string `json:"service_name,omitempty" example:"Yandex Plus"`
	Price       *int    `json:"price,omitempty" example:"450"`
	StartDate   *string `json:"start_date,omitempty" example:"08-2025"`
	EndDate     *string `json:"end_date,omitempty" example:""` // "" — очистить конец
}

// TotalCostResponse — суммарная стоимость подписок
type TotalCostResponse struct {
	Total int `json:"total"`
}
