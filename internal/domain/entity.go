package domain

import (
	"time"

	"github.com/google/uuid"
)

// AdministrativeStructure represents municipal hierarchy nodes (City, Sub-City, Woreda)
type AdministrativeStructure struct {
	ID        uint                    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string                  `gorm:"size:100;not null" json:"name"`
	Level     AdminLevel              `gorm:"size:20;not null;index" json:"level"`
	ParentID  *uint                   `gorm:"index" json:"parent_id,omitempty"`
	Parent    *AdministrativeStructure `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children  []AdministrativeStructure `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Code      string                  `gorm:"size:20;uniqueIndex;not null" json:"code"`
	CreatedAt time.Time               `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

// User represents municipal officers, directors, supervisors and admins
type User struct {
	ID           uuid.UUID                `gorm:"type:uuid;primaryKey" json:"id"`
	FullName     string                   `gorm:"size:150;not null" json:"full_name"`
	Email        string                   `gorm:"size:100;uniqueIndex;not null" json:"email"`
	PhoneNumber  string                   `gorm:"size:20;uniqueIndex;not null" json:"phone_number"`
	PasswordHash string                   `gorm:"size:255;not null" json:"-"`
	Role         UserRole                 `gorm:"size:30;not null;index" json:"role"`
	StructureID  *uint                    `gorm:"index" json:"structure_id,omitempty"`
	Structure    *AdministrativeStructure `gorm:"foreignKey:StructureID" json:"structure,omitempty"`
	IsActive     bool                     `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time                `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

// Citizen represents the public service applicant
type Citizen struct {
	ID                uuid.UUID                `gorm:"type:uuid;primaryKey" json:"id"`
	FullName          string                   `gorm:"size:150;not null" json:"full_name"`
	PhoneNumber       string                   `gorm:"size:20;uniqueIndex;not null" json:"phone_number"`
	NationalID        string                   `gorm:"size:50" json:"national_id,omitempty"`
	TelegramChatID    *int64                   `gorm:"uniqueIndex" json:"telegram_chat_id,omitempty"`
	WoredaID          *uint                    `gorm:"index" json:"woreda_id,omitempty"`
	Woreda            *AdministrativeStructure `gorm:"foreignKey:WoredaID" json:"woreda,omitempty"`
	HouseNumber       string                   `gorm:"size:50" json:"house_number,omitempty"`
	PreferredLanguage string                   `gorm:"size:10;default:'am'" json:"preferred_language"`
	CreatedAt         time.Time                `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

// ServiceType defines public service categories with default SLA response targets
type ServiceType struct {
	ID                  uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name                string `gorm:"size:100;not null" json:"name"`
	Code                string `gorm:"size:30;uniqueIndex;not null" json:"code"`
	BaseSLAHours        int    `gorm:"not null;default:48" json:"base_sla_hours"`
	RequiresWoredaFirst bool   `gorm:"default:true" json:"requires_woreda_first"`
}

// Case represents a citizen grievance ticket undergoing SLA tracking and escalation
type Case struct {
	ID                 uuid.UUID                `gorm:"type:uuid;primaryKey" json:"id"`
	TicketNumber       string                   `gorm:"size:30;uniqueIndex;not null" json:"ticket_number"`
	CitizenID          uuid.UUID                `gorm:"type:uuid;not null;index" json:"citizen_id"`
	Citizen            *Citizen                 `gorm:"foreignKey:CitizenID" json:"citizen,omitempty"`
	ServiceTypeID      uint                     `gorm:"not null;index" json:"service_type_id"`
	ServiceType        *ServiceType             `gorm:"foreignKey:ServiceTypeID" json:"service_type,omitempty"`
	CurrentStructureID uint                     `gorm:"not null;index" json:"current_structure_id"`
	CurrentStructure   *AdministrativeStructure `gorm:"foreignKey:CurrentStructureID" json:"current_structure,omitempty"`
	AssignedToUserID   *uuid.UUID               `gorm:"type:uuid;index" json:"assigned_to_user_id,omitempty"`
	AssignedToUser     *User                    `gorm:"foreignKey:AssignedToUserID" json:"assigned_to_user,omitempty"`
	Title              string                   `gorm:"size:200;not null" json:"title"`
	Description        string                   `gorm:"type:text;not null" json:"description"`
	Status             CaseStatus               `gorm:"size:30;default:'SUBMITTED';index" json:"status"`
	Priority           Priority                 `gorm:"size:20;default:'NORMAL'" json:"priority"`
	SLADeadline        time.Time                `gorm:"not null;index" json:"sla_deadline"`
	WorkingDaysRemaining float64                `gorm:"default:0" json:"working_days_remaining"`
	IsEscalated        bool                     `gorm:"default:false;index" json:"is_escalated"`
	EscalationCount    int                      `gorm:"default:0" json:"escalation_count"`
	ResolutionSummary  string                   `gorm:"type:text" json:"resolution_summary,omitempty"`
	ResolvedBy         *uuid.UUID               `gorm:"type:uuid" json:"resolved_by,omitempty"`
	ResolvedByUser     *User                    `gorm:"foreignKey:ResolvedBy" json:"resolved_by_user,omitempty"`
	ResolvedAt         *time.Time               `json:"resolved_at,omitempty"`
	QRVerificationCode string                   `gorm:"size:100;uniqueIndex" json:"qr_verification_code"`
	Attachments        []CaseAttachment         `gorm:"foreignKey:CaseID;constraint:OnDelete:CASCADE" json:"attachments,omitempty"`
	HearingSlots       []HearingSlot            `gorm:"foreignKey:CaseID;constraint:OnDelete:CASCADE" json:"hearing_slots,omitempty"`
	AuditLogs          []CaseAuditLog           `gorm:"foreignKey:CaseID;constraint:OnDelete:CASCADE" json:"audit_logs,omitempty"`
	CreatedAt          time.Time                `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          time.Time                `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// CaseAttachment stores evidence documents, voice memos, and extracted OCR text
type CaseAttachment struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CaseID           uuid.UUID `gorm:"type:uuid;not null;index" json:"case_id"`
	FileName         string    `gorm:"size:255;not null" json:"file_name"`
	FileURL          string    `gorm:"type:text;not null" json:"file_url"`
	MIMEType         string    `gorm:"size:50;not null" json:"mime_type"`
	ExtractedOCRText string    `gorm:"type:text" json:"extracted_ocr_text,omitempty"`
	UploadedAt       time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"uploaded_at"`
}

// HearingSlot represents scheduled Wednesday/Friday digital or physical hearing sessions
type HearingSlot struct {
	ID           uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	CaseID       uuid.UUID     `gorm:"type:uuid;not null;index" json:"case_id"`
	Case         *Case         `gorm:"foreignKey:CaseID" json:"case,omitempty"`
	OfficialID   uuid.UUID     `gorm:"type:uuid;not null;index" json:"official_id"`
	Official     *User         `gorm:"foreignKey:OfficialID" json:"official,omitempty"`
	HearingDate  time.Time     `gorm:"type:date;not null;index" json:"hearing_date"`
	StartTime    string        `gorm:"size:10;not null" json:"start_time"` // e.g. "09:30"
	EndTime      string        `gorm:"size:10;not null" json:"end_time"`   // e.g. "09:45"
	HearingType  HearingType   `gorm:"size:30;default:'VIRTUAL_CALL'" json:"hearing_type"`
	MeetingLink  string        `gorm:"size:300" json:"meeting_link,omitempty"`
	Status       HearingStatus `gorm:"size:30;default:'SCHEDULED'" json:"status"`
	HearingNotes string        `gorm:"type:text" json:"hearing_notes,omitempty"`
	CreatedAt    time.Time     `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

// CaseAuditLog represents the immutable append-only ledger of every case mutation
type CaseAuditLog struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CaseID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"case_id"`
	PerformedBy *uuid.UUID `gorm:"type:uuid;index" json:"performed_by,omitempty"`
	Performer   *User      `gorm:"foreignKey:PerformedBy" json:"performer,omitempty"`
	Action      string     `gorm:"size:100;not null" json:"action"`
	OldValue    string     `gorm:"type:text" json:"old_value,omitempty"` // JSON string representation
	NewValue    string     `gorm:"type:text" json:"new_value,omitempty"` // JSON string representation
	Notes       string     `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt   time.Time  `gorm:"default:CURRENT_TIMESTAMP;index" json:"created_at"`
}

// EscalationPolicy defines cross-tier SLA threshold hours
type EscalationPolicy struct {
	ID                     uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	ServiceTypeID          *uint      `json:"service_type_id,omitempty"`
	FromLevel              AdminLevel `gorm:"size:20;not null" json:"from_level"`
	ToLevel                AdminLevel `gorm:"size:20;not null" json:"to_level"`
	MaxResolutionHours     int        `gorm:"not null" json:"max_resolution_hours"`
	WarningThresholdHours int        `gorm:"not null" json:"warning_threshold_hours"`
	IsActive               bool       `gorm:"default:true" json:"is_active"`
}

// Holiday represents Ethiopian public, national and religious holidays
type Holiday struct {
	ID          uint                     `gorm:"primaryKey;autoIncrement" json:"id"`
	Date        time.Time                `gorm:"type:date;not null;index" json:"date"`
	NameEN      string                   `gorm:"size:150;not null" json:"name_en"`
	NameAM      string                   `gorm:"size:150;not null" json:"name_am"`
	Type        HolidayType              `gorm:"size:20;not null;index" json:"type"` // FIXED, MOVABLE, ISLAMIC
	Active      bool                     `gorm:"default:true;index" json:"active"`
	Source      string                   `gorm:"size:100" json:"source"`
	Year        int                      `gorm:"not null;index" json:"year"`
	StructureID *uint                    `gorm:"index" json:"structure_id,omitempty"`
	Structure   *AdministrativeStructure `gorm:"foreignKey:StructureID" json:"structure,omitempty"`
	CreatedAt   time.Time                `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time                `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}
