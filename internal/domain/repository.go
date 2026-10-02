package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// CaseFilter parameters for querying tickets
type CaseFilter struct {
	StructureID     *uint
	StructureIDs    []uint // Scoped list of authorized administrative structures (Sub-City, Woredas)
	CitizenID       *uuid.UUID
	AssignedToID    *uuid.UUID
	Status          *CaseStatus
	Priority        *Priority
	IsEscalated     *bool
	OnlyBreached    bool
	SearchQuery     string
	Limit           int
	Offset          int
}

// CaseRepository defines the persistence port for cases
type CaseRepository interface {
	Create(ctx context.Context, c *Case) error
	GetByID(ctx context.Context, id uuid.UUID) (*Case, error)
	GetByTicketNumber(ctx context.Context, ticket string) (*Case, error)
	GetByQRCode(ctx context.Context, qrCode string) (*Case, error)
	Update(ctx context.Context, c *Case) error
	List(ctx context.Context, filter CaseFilter) ([]Case, int64, error)
	FetchAndLockBreachedCases(ctx context.Context, now time.Time, limit int) ([]Case, error)
	CountByStatus(ctx context.Context, structureID *uint) (map[CaseStatus]int64, error)
	AddAttachment(ctx context.Context, att *CaseAttachment) error
}

// StructureRepository defines administrative structure hierarchy lookups
type StructureRepository interface {
	Create(ctx context.Context, s *AdministrativeStructure) error
	GetByID(ctx context.Context, id uint) (*AdministrativeStructure, error)
	GetByCode(ctx context.Context, code string) (*AdministrativeStructure, error)
	ListAll(ctx context.Context) ([]AdministrativeStructure, error)
	GetChildren(ctx context.Context, parentID uint) ([]AdministrativeStructure, error)
	GetHierarchyTree(ctx context.Context) ([]AdministrativeStructure, error)
}

// UserRepository defines user persistence operations
type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	ListByStructure(ctx context.Context, structureID uint) ([]User, error)
	ListAll(ctx context.Context) ([]User, error)
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// CitizenRepository defines citizen profile lookup and registration
type CitizenRepository interface {
	Create(ctx context.Context, c *Citizen) error
	GetByID(ctx context.Context, id uuid.UUID) (*Citizen, error)
	GetByPhone(ctx context.Context, phone string) (*Citizen, error)
	GetByTelegramID(ctx context.Context, chatID int64) (*Citizen, error)
}

// HearingRepository defines Wednesday/Friday hearing desks
type HearingRepository interface {
	CreateSlot(ctx context.Context, slot *HearingSlot) error
	GetByID(ctx context.Context, id uuid.UUID) (*HearingSlot, error)
	GetByCaseID(ctx context.Context, caseID uuid.UUID) ([]HearingSlot, error)
	ListByDate(ctx context.Context, date time.Time, officialID *uuid.UUID) ([]HearingSlot, error)
	UpdateSlot(ctx context.Context, slot *HearingSlot) error
}

// AuditRepository defines the immutable append-only ledger port
type AuditRepository interface {
	Log(ctx context.Context, log *CaseAuditLog) error
	GetByCaseID(ctx context.Context, caseID uuid.UUID) ([]CaseAuditLog, error)
	ListRecent(ctx context.Context, limit int) ([]CaseAuditLog, error)
}

// ServiceTypeRepository defines access to grievance classifications and default SLA targets
type ServiceTypeRepository interface {
	GetByID(ctx context.Context, id uint) (*ServiceType, error)
	GetByCode(ctx context.Context, code string) (*ServiceType, error)
	ListAll(ctx context.Context) ([]ServiceType, error)
	Create(ctx context.Context, st *ServiceType) error
}

// HolidayRepository defines holiday persistence operations
type HolidayRepository interface {
	Create(ctx context.Context, h *Holiday) error
	GetByID(ctx context.Context, id uint) (*Holiday, error)
	GetByDate(ctx context.Context, date time.Time) (*Holiday, error)
	ListByYear(ctx context.Context, year int) ([]Holiday, error)
	ListAll(ctx context.Context) ([]Holiday, error)
	Update(ctx context.Context, h *Holiday) error
	Delete(ctx context.Context, id uint) error
	IsHoliday(ctx context.Context, date time.Time, structureID *uint) (bool, error)
}
