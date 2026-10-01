package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/smart-onestop/platform/internal/domain"
	"gorm.io/gorm"
)

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Create(ctx context.Context, u *domain.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *GormUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).Preload("Structure").First(&u, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *GormUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).Preload("Structure").First(&u, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *GormUserRepository) ListByStructure(ctx context.Context, structureID uint) ([]domain.User, error) {
	var users []domain.User
	err := r.db.WithContext(ctx).Where("structure_id = ? AND is_active = true", structureID).Find(&users).Error
	return users, err
}

func (r *GormUserRepository) ListAll(ctx context.Context) ([]domain.User, error) {
	var users []domain.User
	err := r.db.WithContext(ctx).Preload("Structure").Order("created_at desc").Find(&users).Error
	return users, err
}

func (r *GormUserRepository) Update(ctx context.Context, u *domain.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *GormUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.User{}, "id = ?", id).Error
}
