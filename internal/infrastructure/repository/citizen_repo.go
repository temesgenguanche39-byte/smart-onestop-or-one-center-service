package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/smart-onestop/platform/internal/domain"
	"gorm.io/gorm"
)

type GormCitizenRepository struct {
	db *gorm.DB
}

func NewGormCitizenRepository(db *gorm.DB) *GormCitizenRepository {
	return &GormCitizenRepository{db: db}
}

func (r *GormCitizenRepository) Create(ctx context.Context, c *domain.Citizen) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *GormCitizenRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Citizen, error) {
	var c domain.Citizen
	err := r.db.WithContext(ctx).Preload("Woreda").First(&c, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *GormCitizenRepository) GetByPhone(ctx context.Context, phone string) (*domain.Citizen, error) {
	var c domain.Citizen
	err := r.db.WithContext(ctx).Preload("Woreda").First(&c, "phone_number = ?", phone).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *GormCitizenRepository) GetByTelegramID(ctx context.Context, chatID int64) (*domain.Citizen, error) {
	var c domain.Citizen
	err := r.db.WithContext(ctx).Preload("Woreda").First(&c, "telegram_chat_id = ?", chatID).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}
