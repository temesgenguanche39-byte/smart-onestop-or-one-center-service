package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/smart-onestop/platform/internal/domain"
	"gorm.io/gorm"
)

type GormAuditRepository struct {
	db *gorm.DB
}

func NewGormAuditRepository(db *gorm.DB) *GormAuditRepository {
	return &GormAuditRepository{db: db}
}

// Log writes an immutable audit record to the ledger
func (r *GormAuditRepository) Log(ctx context.Context, log *domain.CaseAuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *GormAuditRepository) GetByCaseID(ctx context.Context, caseID uuid.UUID) ([]domain.CaseAuditLog, error) {
	var logs []domain.CaseAuditLog
	err := r.db.WithContext(ctx).
		Preload("Performer").
		Where("case_id = ?", caseID).
		Order("created_at ASC").
		Find(&logs).Error
	return logs, err
}

func (r *GormAuditRepository) ListRecent(ctx context.Context, limit int) ([]domain.CaseAuditLog, error) {
	var logs []domain.CaseAuditLog
	if limit <= 0 {
		limit = 20
	}
	err := r.db.WithContext(ctx).
		Preload("Performer").
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}
