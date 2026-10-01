package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/smart-onestop/platform/internal/domain"
)

// CreateCaseRequest received from Citizen Intake Web/PWA/Telegram
type CreateCaseRequest struct {
	CitizenFullName    string          `json:"citizen_full_name" binding:"required"`
	CitizenPhone       string          `json:"citizen_phone" binding:"required"`
	CitizenNationalID  string          `json:"citizen_national_id"`
	CitizenHouseNumber string          `json:"citizen_house_number"`
	WoredaID           uint            `json:"woreda_id" binding:"required"`
	ServiceTypeID      uint            `json:"service_type_id" binding:"required"`
	Title              string          `json:"title" binding:"required"`
	Description        string          `json:"description" binding:"required"`
	Priority           domain.Priority `json:"priority"`
	Attachments        []AttachmentDTO `json:"attachments"`
}

type AttachmentDTO struct {
	FileName         string `json:"file_name"`
	FileURL          string `json:"file_url"`
	MIMEType         string `json:"mime_type"`
	ExtractedOCRText string `json:"extracted_ocr_text"`
}

type AddAttachmentRequest struct {
	FileName         string `json:"file_name" binding:"required"`
	FileURL          string `json:"file_url" binding:"required"`
	MIMEType         string `json:"mime_type"`
	ExtractedOCRText string `json:"extracted_ocr_text"`
}

// CaseResponse representation for listing
type CaseResponse struct {
	ID                 uuid.UUID               `json:"id"`
	TicketNumber       string                  `json:"ticket_number"`
	CitizenName        string                  `json:"citizen_name"`
	CitizenPhone       string                  `json:"citizen_phone"`
	ServiceTypeName    string                  `json:"service_type_name"`
	CurrentStructureID uint                    `json:"current_structure_id"`
	CurrentStructure   string                  `json:"current_structure_name"`
	StructureLevel     domain.AdminLevel       `json:"structure_level"`
	AssignedToName     string                  `json:"assigned_to_name,omitempty"`
	Title              string                  `json:"title"`
	Status             domain.CaseStatus       `json:"status"`
	Priority           domain.Priority         `json:"priority"`
	SLADeadline        time.Time               `json:"sla_deadline"`
	RemainingHours     float64                 `json:"remaining_hours"`
	IsBreached         bool                    `json:"is_breached"`
	IsEscalated        bool                    `json:"is_escalated"`
	EscalationCount    int                     `json:"escalation_count"`
	QRVerificationCode string                  `json:"qr_verification_code"`
	CreatedAt          time.Time               `json:"created_at"`
}

// CaseDetailResponse includes attachments, hearings, and immutable audit logs
type CaseDetailResponse struct {
	CaseResponse
	Description       string                   `json:"description"`
	ResolutionSummary string                   `json:"resolution_summary,omitempty"`
	ResolvedByName    string                   `json:"resolved_by_name,omitempty"`
	ResolvedAt        *time.Time               `json:"resolved_at,omitempty"`
	Attachments       []domain.CaseAttachment  `json:"attachments"`
	HearingSlots      []HearingResponse        `json:"hearing_slots"`
	AuditLogs         []AuditLogResponse       `json:"audit_logs"`
}

// ResolveCaseRequest submitted by municipal decision-maker
type ResolveCaseRequest struct {
	ResolutionSummary string `json:"resolution_summary" binding:"required"`
	DecisionNotes     string `json:"decision_notes"`
	SendNotification  bool   `json:"send_notification"`
}

// AssignCaseRequest assigned to specific municipal officer
type AssignCaseRequest struct {
	AssignedToUserID uuid.UUID `json:"assigned_to_user_id" binding:"required"`
	Notes            string    `json:"notes"`
}

// ManualEscalateRequest when officer escalates before deadline
type ManualEscalateRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// ScheduleHearingRequest for Wednesday / Friday desk sessions
type ScheduleHearingRequest struct {
	HearingDate string             `json:"hearing_date" binding:"required"` // Format: YYYY-MM-DD
	StartTime   string             `json:"start_time" binding:"required"`   // e.g. "09:30"
	EndTime     string             `json:"end_time" binding:"required"`     // e.g. "09:45"
	HearingType domain.HearingType `json:"hearing_type"`
	MeetingLink string             `json:"meeting_link"`
	Notes       string             `json:"notes"`
}

// HearingResponse DTO
type HearingResponse struct {
	ID           uuid.UUID            `json:"id"`
	CaseID       uuid.UUID            `json:"case_id"`
	TicketNumber string               `json:"ticket_number,omitempty"`
	OfficialID   uuid.UUID            `json:"official_id"`
	OfficialName string               `json:"official_name"`
	HearingDate  string               `json:"hearing_date"`
	StartTime    string               `json:"start_time"`
	EndTime      string               `json:"end_time"`
	HearingType  domain.HearingType   `json:"hearing_type"`
	MeetingLink  string               `json:"meeting_link"`
	Status       domain.HearingStatus `json:"status"`
	HearingNotes string               `json:"hearing_notes,omitempty"`
}

// AuditLogResponse DTO
type AuditLogResponse struct {
	ID            uint64    `json:"id"`
	CaseID        uuid.UUID `json:"case_id"`
	Action        string    `json:"action"`
	PerformerName string    `json:"performer_name"`
	OldValue      string    `json:"old_value,omitempty"`
	NewValue      string    `json:"new_value,omitempty"`
	Notes         string    `json:"notes,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// Auth Requests & Responses
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token        string           `json:"token"`
	User         UserDTO          `json:"user"`
	ExpiresAt    time.Time        `json:"expires_at"`
}

type UserDTO struct {
	ID            uuid.UUID         `json:"id"`
	FullName      string            `json:"full_name"`
	Email         string            `json:"email"`
	PhoneNumber   string            `json:"phone_number"`
	Role          domain.UserRole   `json:"role"`
	StructureID   *uint             `json:"structure_id,omitempty"`
	StructureName string            `json:"structure_name,omitempty"`
	AdminLevel    domain.AdminLevel `json:"admin_level,omitempty"`
}

type CreateUserRequest struct {
	FullName    string          `json:"full_name" binding:"required"`
	Email       string          `json:"email" binding:"required,email"`
	PhoneNumber string          `json:"phone_number" binding:"required"`
	Password    string          `json:"password" binding:"required,min=6"`
	Role        domain.UserRole `json:"role" binding:"required"`
	StructureID *uint           `json:"structure_id"`
}

// AnalyticsSummaryResponse for supervisor & city director command center
type AnalyticsSummaryResponse struct {
	TotalCases           int64                    `json:"total_cases"`
	ActiveCases          int64                    `json:"active_cases"`
	ResolvedCases        int64                    `json:"resolved_cases"`
	UnderReviewCases     int64                    `json:"under_review_cases"`
	ScheduledHearings    int64                    `json:"scheduled_hearings"`
	BreachedCases        int64                    `json:"breached_cases"`
	EscalatedToSubCity   int64                    `json:"escalated_to_subcity"`
	EscalatedToCity      int64                    `json:"escalated_to_city"`
	AverageResolutionHrs float64                  `json:"avg_resolution_hours"`
	StructureBreakdown   []StructureMetric        `json:"structure_breakdown"`
	RecentAuditLogs      []AuditLogResponse       `json:"recent_audit_logs"`
}

type StructureMetric struct {
	StructureID   uint              `json:"structure_id"`
	Name          string            `json:"name"`
	Level         domain.AdminLevel `json:"level"`
	CaseCount     int64             `json:"case_count"`
	BreachedCount int64             `json:"breached_count"`
	ResolvedCount int64             `json:"resolved_count"`
}
