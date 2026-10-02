package repository

import (
	"context"
	"time"

	"github.com/smart-onestop/platform/internal/domain"
	"gorm.io/gorm"
)

type GormHolidayRepository struct {
	db *gorm.DB
}

func NewGormHolidayRepository(db *gorm.DB) *GormHolidayRepository {
	return &GormHolidayRepository{db: db}
}

func (r *GormHolidayRepository) Create(ctx context.Context, h *domain.Holiday) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *GormHolidayRepository) GetByID(ctx context.Context, id uint) (*domain.Holiday, error) {
	var h domain.Holiday
	err := r.db.WithContext(ctx).First(&h, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrHolidayNotFound
		}
		return nil, err
	}
	return &h, nil
}

func (r *GormHolidayRepository) GetByDate(ctx context.Context, date time.Time) (*domain.Holiday, error) {
	var h domain.Holiday
	dateStr := date.Format("2006-01-02")
	err := r.db.WithContext(ctx).Where("date = ? AND active = ?", dateStr, true).First(&h).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrHolidayNotFound
		}
		return nil, err
	}
	return &h, nil
}

func (r *GormHolidayRepository) ListByYear(ctx context.Context, year int) ([]domain.Holiday, error) {
	var holidays []domain.Holiday
	query := r.db.WithContext(ctx).Where("active = ?", true)
	if year > 0 {
		query = query.Where("year = ?", year)
	}
	err := query.Order("date ASC").Find(&holidays).Error
	return holidays, err
}

func (r *GormHolidayRepository) ListAll(ctx context.Context) ([]domain.Holiday, error) {
	var holidays []domain.Holiday
	err := r.db.WithContext(ctx).Order("date ASC").Find(&holidays).Error
	return holidays, err
}

func (r *GormHolidayRepository) Update(ctx context.Context, h *domain.Holiday) error {
	return r.db.WithContext(ctx).Save(h).Error
}

func (r *GormHolidayRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Holiday{}, id).Error
}

func (r *GormHolidayRepository) IsHoliday(ctx context.Context, date time.Time, structureID *uint) (bool, error) {
	dateStr := date.Format("2006-01-02")
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.Holiday{}).
		Where("date = ? AND active = ?", dateStr, true)
	if structureID != nil && *structureID > 0 {
		query = query.Where("(structure_id IS NULL OR structure_id = ?)", *structureID)
	} else {
		query = query.Where("structure_id IS NULL")
	}
	err := query.Count(&count).Error
	return count > 0, err
}
