package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/smart-onestop/platform/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormCaseRepository struct {
	db *gorm.DB
}

func NewGormCaseRepository(db *gorm.DB) *GormCaseRepository {
	return &GormCaseRepository{db: db}
}

func (r *GormCaseRepository) Create(ctx context.Context, c *domain.Case) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *GormCaseRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Case, error) {
	var c domain.Case
	err := r.db.WithContext(ctx).
		Preload("Citizen").
		Preload("ServiceType").
		Preload("CurrentStructure").
		Preload("AssignedToUser").
		Preload("ResolvedByUser").
		Preload("Attachments").
		First(&c, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *GormCaseRepository) GetByTicketNumber(ctx context.Context, ticket string) (*domain.Case, error) {
	var c domain.Case
	err := r.db.WithContext(ctx).
		Preload("Citizen").
		Preload("ServiceType").
		Preload("CurrentStructure").
		Preload("AssignedToUser").
		Preload("ResolvedByUser").
		Preload("Attachments").
		First(&c, "ticket_number = ?", ticket).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *GormCaseRepository) GetByQRCode(ctx context.Context, qrCode string) (*domain.Case, error) {
	var c domain.Case
	err := r.db.WithContext(ctx).
		Preload("Citizen").
		Preload("ServiceType").
		Preload("CurrentStructure").
		Preload("AssignedToUser").
		Preload("ResolvedByUser").
		Preload("Attachments").
		First(&c, "qr_verification_code = ?", qrCode).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *GormCaseRepository) Update(ctx context.Context, c *domain.Case) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *GormCaseRepository) List(ctx context.Context, filter domain.CaseFilter) ([]domain.Case, int64, error) {
	var cases []domain.Case
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.Case{}).
		Preload("Citizen").
		Preload("ServiceType").
		Preload("CurrentStructure").
		Preload("AssignedToUser")

	if filter.StructureID != nil {
		q = q.Where("current_structure_id = ?", *filter.StructureID)
	} else if len(filter.StructureIDs) > 0 {
		q = q.Where("current_structure_id IN ?", filter.StructureIDs)
	}
	if filter.CitizenID != nil {
		q = q.Where("citizen_id = ?", *filter.CitizenID)
	}
	if filter.AssignedToID != nil {
		q = q.Where("assigned_to_user_id = ?", *filter.AssignedToID)
	}
	if filter.Status != nil {
		q = q.Where("status = ?", *filter.Status)
	}
	if filter.Priority != nil {
		q = q.Where("priority = ?", *filter.Priority)
	}
	if filter.IsEscalated != nil {
		q = q.Where("is_escalated = ?", *filter.IsEscalated)
	}
	if filter.OnlyBreached {
		q = q.Where("sla_deadline <= ? AND status NOT IN ('RESOLVED', 'REJECTED')", time.Now().UTC())
	}
	if filter.SearchQuery != "" {
		like := "%" + filter.SearchQuery + "%"
		q = q.Where("ticket_number LIKE ? OR title LIKE ?", like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset

	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&cases).Error; err != nil {
		return nil, 0, err
	}

	return cases, total, nil
}

// FetchAndLockBreachedCases implements row-level locking (FOR UPDATE SKIP LOCKED)
func (r *GormCaseRepository) FetchAndLockBreachedCases(ctx context.Context, now time.Time, limit int) ([]domain.Case, error) {
	var breachedCases []domain.Case

	// Check if dialect is postgres, add SKIP LOCKED
	tx := r.db.WithContext(ctx)
	if r.db.Dialector.Name() == "postgres" {
		tx = tx.Clauses(clause.Locking{
			Strength: "UPDATE",
			Table:    clause.Table{Name: clause.CurrentTable},
			Options:  "SKIP LOCKED",
		})
	}

	err := tx.
		Where("status NOT IN (?, ?)", domain.StatusResolved, domain.StatusRejected).
		Where("sla_deadline <= ?", now).
		Limit(limit).
		Find(&breachedCases).Error

	return breachedCases, err
}

func (r *GormCaseRepository) CountByStatus(ctx context.Context, structureID *uint) (map[domain.CaseStatus]int64, error) {
	type StatusCount struct {
		Status domain.CaseStatus
		Count  int64
	}
	var results []StatusCount

	q := r.db.WithContext(ctx).Model(&domain.Case{}).
		Select("status, count(*) as count")

	if structureID != nil {
		q = q.Where("current_structure_id = ?", *structureID)
	}

	err := q.Group("status").Scan(&results).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[domain.CaseStatus]int64)
	for _, r := range results {
		counts[r.Status] = r.Count
	}
	return counts, nil
}

func (r *GormCaseRepository) AddAttachment(ctx context.Context, att *domain.CaseAttachment) error {
	return r.db.WithContext(ctx).Create(att).Error
}
