package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/smart-onestop/platform/internal/domain"
	"gorm.io/gorm"
)

type GormHearingRepository struct {
	db *gorm.DB
}

func NewGormHearingRepository(db *gorm.DB) *GormHearingRepository {
	return &GormHearingRepository{db: db}
}

func (r *GormHearingRepository) CreateSlot(ctx context.Context, slot *domain.HearingSlot) error {
	return r.db.WithContext(ctx).Create(slot).Error
}

func (r *GormHearingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.HearingSlot, error) {
	var slot domain.HearingSlot
	err := r.db.WithContext(ctx).
		Preload("Official").
		Preload("Case").
		First(&slot, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &slot, nil
}

func (r *GormHearingRepository) GetByCaseID(ctx context.Context, caseID uuid.UUID) ([]domain.HearingSlot, error) {
	var slots []domain.HearingSlot
	err := r.db.WithContext(ctx).
		Preload("Official").
		Where("case_id = ?", caseID).
		Order("hearing_date ASC, start_time ASC").
		Find(&slots).Error
	return slots, err
}

func (r *GormHearingRepository) ListByDate(ctx context.Context, date time.Time, officialID *uuid.UUID) ([]domain.HearingSlot, error) {
	var slots []domain.HearingSlot
	q := r.db.WithContext(ctx).Preload("Official").Preload("Case")

	// Match date
	q = q.Where("hearing_date = ?", date)

	if officialID != nil {
		q = q.Where("official_id = ?", *officialID)
	}

	err := q.Order("start_time ASC").Find(&slots).Error
	return slots, err
}

func (r *GormHearingRepository) UpdateSlot(ctx context.Context, slot *domain.HearingSlot) error {
	return r.db.WithContext(ctx).Save(slot).Error
}
