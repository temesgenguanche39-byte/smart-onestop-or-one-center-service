package repository

import (
	"context"

	"github.com/smart-onestop/platform/internal/domain"
	"gorm.io/gorm"
)

type GormStructureRepository struct {
	db *gorm.DB
}

func NewGormStructureRepository(db *gorm.DB) *GormStructureRepository {
	return &GormStructureRepository{db: db}
}

func (r *GormStructureRepository) Create(ctx context.Context, s *domain.AdministrativeStructure) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *GormStructureRepository) GetByID(ctx context.Context, id uint) (*domain.AdministrativeStructure, error) {
	var s domain.AdministrativeStructure
	err := r.db.WithContext(ctx).Preload("Parent").First(&s, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *GormStructureRepository) GetByCode(ctx context.Context, code string) (*domain.AdministrativeStructure, error) {
	var s domain.AdministrativeStructure
	err := r.db.WithContext(ctx).First(&s, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *GormStructureRepository) ListAll(ctx context.Context) ([]domain.AdministrativeStructure, error) {
	var list []domain.AdministrativeStructure
	err := r.db.WithContext(ctx).Order("level ASC, name ASC").Find(&list).Error
	return list, err
}

func (r *GormStructureRepository) GetChildren(ctx context.Context, parentID uint) ([]domain.AdministrativeStructure, error) {
	var children []domain.AdministrativeStructure
	err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Order("name ASC").Find(&children).Error
	return children, err
}

func (r *GormStructureRepository) GetHierarchyTree(ctx context.Context) ([]domain.AdministrativeStructure, error) {
	var roots []domain.AdministrativeStructure
	err := r.db.WithContext(ctx).
		Where("parent_id IS NULL").
		Preload("Children.Children").
		Find(&roots).Error
	return roots, err
}
