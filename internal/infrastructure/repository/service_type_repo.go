package repository

import (
	"context"

	"github.com/smart-onestop/platform/internal/domain"
	"gorm.io/gorm"
)

type GormServiceTypeRepository struct {
	db *gorm.DB
}

func NewGormServiceTypeRepository(db *gorm.DB) *GormServiceTypeRepository {
	return &GormServiceTypeRepository{db: db}
}

func (r *GormServiceTypeRepository) GetByID(ctx context.Context, id uint) (*domain.ServiceType, error) {
	var st domain.ServiceType
	err := r.db.WithContext(ctx).First(&st, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &st, nil
}

func (r *GormServiceTypeRepository) GetByCode(ctx context.Context, code string) (*domain.ServiceType, error) {
	var st domain.ServiceType
	err := r.db.WithContext(ctx).First(&st, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &st, nil
}

func (r *GormServiceTypeRepository) ListAll(ctx context.Context) ([]domain.ServiceType, error) {
	var list []domain.ServiceType
	err := r.db.WithContext(ctx).Order("name ASC").Find(&list).Error
	return list, err
}

func (r *GormServiceTypeRepository) Create(ctx context.Context, st *domain.ServiceType) error {
	return r.db.WithContext(ctx).Create(st).Error
}
