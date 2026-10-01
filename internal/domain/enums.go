package domain

// AdminLevel represents the hierarchical administrative governance tiers
type AdminLevel string

const (
	AdminLevelCity    AdminLevel = "CITY"
	AdminLevelSubCity AdminLevel = "SUB_CITY"
	AdminLevelWoreda  AdminLevel = "WOREDA"
)

// UserRole represents declarative RBAC permissions
type UserRole string

const (
	RoleSuperAdmin       UserRole = "SUPER_ADMIN"
	RoleCityDirector     UserRole = "CITY_DIRECTOR"
	RoleSubcityManager   UserRole = "SUBCITY_MANAGER"
	RoleWoredaOfficer    UserRole = "WOREDA_OFFICER"
	RoleServiceDeskAgent UserRole = "SERVICE_DESK_AGENT"
	RoleCitizen          UserRole = "CITIZEN"
)

// CaseStatus represents the deterministic lifecycle state of a grievance
type CaseStatus string

const (
	StatusSubmitted           CaseStatus = "SUBMITTED"
	StatusUnderReview         CaseStatus = "UNDER_REVIEW"
	StatusScheduledForHearing CaseStatus = "SCHEDULED_FOR_HEARING"
	StatusResolved            CaseStatus = "RESOLVED"
	StatusRejected            CaseStatus = "REJECTED"
	StatusEscalatedToSubCity  CaseStatus = "ESCALATED_TO_SUBCITY"
	StatusEscalatedToCity     CaseStatus = "ESCALATED_TO_CITY"
)

// Priority represents the urgency of the case
type Priority string

const (
	PriorityLow    Priority = "LOW"
	PriorityNormal Priority = "NORMAL"
	PriorityHigh   Priority = "HIGH"
	PriorityUrgent Priority = "URGENT"
)

// HearingType denotes virtual or physical hearing
type HearingType string

const (
	HearingTypeVirtualCall HearingType = "VIRTUAL_CALL"
	HearingTypeInPerson    HearingType = "IN_PERSON"
)

// HearingStatus represents the appointment status
type HearingStatus string

const (
	HearingStatusScheduled           HearingStatus = "SCHEDULED"
	HearingStatusCompleted           HearingStatus = "COMPLETED"
	HearingStatusMissedByCitizen     HearingStatus = "MISSED_BY_CITIZEN"
	HearingStatusCancelledByOfficial HearingStatus = "CANCELLED_BY_OFFICIAL"
)
