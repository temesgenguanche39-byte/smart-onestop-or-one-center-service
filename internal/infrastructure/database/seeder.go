package database

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/smart-onestop/platform/internal/domain"
	"github.com/smart-onestop/platform/internal/infrastructure/security"
	"gorm.io/gorm"
)

// SeedDatabase inserts initial municipal hierarchy, service types, users, and demonstration cases
func SeedDatabase(db *gorm.DB) error {
	var count int64
	db.Model(&domain.AdministrativeStructure{}).Count(&count)
	if count > 0 {
		log.Println("[DB Seeder] Database already populated with seed data.")
		return nil
	}

	log.Println("[DB Seeder] Populating database with Ethiopian Municipal administrative hierarchy...")

	// 1. Administrative Hierarchy
	city := domain.AdministrativeStructure{
		Name:      "Addis Ababa City Administration",
		Level:     domain.AdminLevelCity,
		Code:      "ETH-AA-CITY",
		CreatedAt: time.Now().UTC(),
	}
	db.Create(&city)

	kirkos := domain.AdministrativeStructure{
		Name:      "Kirkos Sub-City Administration",
		Level:     domain.AdminLevelSubCity,
		ParentID:  &city.ID,
		Code:      "ETH-AA-KIRKOS",
		CreatedAt: time.Now().UTC(),
	}
	bole := domain.AdministrativeStructure{
		Name:      "Bole Sub-City Administration",
		Level:     domain.AdminLevelSubCity,
		ParentID:  &city.ID,
		Code:      "ETH-AA-BOLE",
		CreatedAt: time.Now().UTC(),
	}
	db.Create(&kirkos)
	db.Create(&bole)

	woreda01 := domain.AdministrativeStructure{
		Name:      "Kirkos Woreda 01 Administration",
		Level:     domain.AdminLevelWoreda,
		ParentID:  &kirkos.ID,
		Code:      "ETH-AA-KIR-W01",
		CreatedAt: time.Now().UTC(),
	}
	woreda02 := domain.AdministrativeStructure{
		Name:      "Kirkos Woreda 02 Administration",
		Level:     domain.AdminLevelWoreda,
		ParentID:  &kirkos.ID,
		Code:      "ETH-AA-KIR-W02",
		CreatedAt: time.Now().UTC(),
	}
	woreda03 := domain.AdministrativeStructure{
		Name:      "Bole Woreda 03 Administration",
		Level:     domain.AdminLevelWoreda,
		ParentID:  &bole.ID,
		Code:      "ETH-AA-BOL-W03",
		CreatedAt: time.Now().UTC(),
	}
	db.Create(&woreda01)
	db.Create(&woreda02)
	db.Create(&woreda03)

	// 2. Service Types
	serviceTypes := []domain.ServiceType{
		{Name: "Land Administration & Title Disputes (የመሬት ይዞታና ካርታ)", Code: "LAND_ADMIN", BaseSLAHours: 48, RequiresWoredaFirst: true},
		{Name: "Trade License & Commercial Registration (የንግድ ፈቃድና ምዝገባ)", Code: "TRADE_LICENSE", BaseSLAHours: 24, RequiresWoredaFirst: true},
		{Name: "Public Housing & Kebele Housing Disputes (የመንግስት ቤቶች ቅሬታ)", Code: "PUBLIC_HOUSING", BaseSLAHours: 72, RequiresWoredaFirst: true},
		{Name: "Vital Events & Civil Registration (የወሳኝ ኩነቶች ምዝገባ)", Code: "VITAL_EVENTS", BaseSLAHours: 24, RequiresWoredaFirst: true},
		{Name: "Municipal Infrastructure & Utilities (የመሰረተ ልማትና መንገድ)", Code: "INFRASTRUCTURE", BaseSLAHours: 48, RequiresWoredaFirst: true},
	}
	for i := range serviceTypes {
		db.Create(&serviceTypes[i])
	}

	// 3. Official Users
	defaultHash, _ := security.HashPassword("Password123!")

	users := []domain.User{
		{
			ID:           uuid.New(),
			FullName:     "Super User",
			Email:        "superuser@smartonestop.gov.et",
			PhoneNumber:  "+251911999999",
			PasswordHash: defaultHash,
			Role:         domain.RoleSuperAdmin,
			StructureID:  &city.ID,
			IsActive:     true,
			CreatedAt:    time.Now().UTC(),
		},
		{
			ID:           uuid.New(),
			FullName:     "Alemayehu Tadesse",
			Email:        "superadmin@smartonestop.gov.et",
			PhoneNumber:  "+251911000001",
			PasswordHash: defaultHash,
			Role:         domain.RoleSuperAdmin,
			StructureID:  &city.ID,
			IsActive:     true,
			CreatedAt:    time.Now().UTC(),
		},
		{
			ID:           uuid.New(),
			FullName:     "Dr. Selamawit Bekele",
			Email:        "city.director@smartonestop.gov.et",
			PhoneNumber:  "+251911000002",
			PasswordHash: defaultHash,
			Role:         domain.RoleCityDirector,
			StructureID:  &city.ID,
			IsActive:     true,
			CreatedAt:    time.Now().UTC(),
		},
		{
			ID:           uuid.New(),
			FullName:     "Kassahun Haile",
			Email:        "kirkos.manager@smartonestop.gov.et",
			PhoneNumber:  "+251911000003",
			PasswordHash: defaultHash,
			Role:         domain.RoleSubcityManager,
			StructureID:  &kirkos.ID,
			IsActive:     true,
			CreatedAt:    time.Now().UTC(),
		},
		{
			ID:           uuid.New(),
			FullName:     "Bethlehem Girma",
			Email:        "woreda01.officer@smartonestop.gov.et",
			PhoneNumber:  "+251911000004",
			PasswordHash: defaultHash,
			Role:         domain.RoleWoredaOfficer,
			StructureID:  &woreda01.ID,
			IsActive:     true,
			CreatedAt:    time.Now().UTC(),
		},
		{
			ID:           uuid.New(),
			FullName:     "Yonas Mulugeta",
			Email:        "desk.agent@smartonestop.gov.et",
			PhoneNumber:  "+251911000005",
			PasswordHash: defaultHash,
			Role:         domain.RoleServiceDeskAgent,
			StructureID:  &woreda01.ID,
			IsActive:     true,
			CreatedAt:    time.Now().UTC(),
		},
	}
	for i := range users {
		db.Create(&users[i])
	}

	woredaOfficer := users[3]

	// 4. Citizens
	citizen1 := domain.Citizen{
		ID:                uuid.New(),
		FullName:          "Tewodros Assefa",
		PhoneNumber:       "+251912345678",
		NationalID:        "ETH-NAT-883492",
		WoredaID:          &woreda01.ID,
		HouseNumber:       "142/B",
		PreferredLanguage: "am",
		CreatedAt:         time.Now().UTC(),
	}
	citizen2 := domain.Citizen{
		ID:                uuid.New(),
		FullName:          "Genet Wolde",
		PhoneNumber:       "+251922334455",
		NationalID:        "ETH-NAT-774411",
		WoredaID:          &woreda01.ID,
		HouseNumber:       "089",
		PreferredLanguage: "am",
		CreatedAt:         time.Now().UTC(),
	}
	citizen3 := domain.Citizen{
		ID:                uuid.New(),
		FullName:          "Dawit Kebede",
		PhoneNumber:       "+251933445566",
		NationalID:        "ETH-NAT-556677",
		WoredaID:          &woreda02.ID,
		HouseNumber:       "311",
		PreferredLanguage: "am",
		CreatedAt:         time.Now().UTC(),
	}
	db.Create(&citizen1)
	db.Create(&citizen2)
	db.Create(&citizen3)

	now := time.Now().UTC()

	// 5. Seed Demonstration Cases
	// Case 1: Active Woreda Review
	case1 := domain.Case{
		ID:                 uuid.New(),
		TicketNumber:       "TKT-2026-100234",
		CitizenID:          citizen1.ID,
		ServiceTypeID:      serviceTypes[0].ID,
		CurrentStructureID: woreda01.ID,
		AssignedToUserID:   &woredaOfficer.ID,
		Title:              "Boundary Wall and Land Survey Discrepancy",
		Description:        "The neighboring commercial plot encroached 2.5 meters onto our certified residential parcel. Woreda survey department hasn't finalized inspection.",
		Status:             domain.StatusUnderReview,
		Priority:           domain.PriorityHigh,
		SLADeadline:        now.Add(36 * time.Hour),
		IsEscalated:        false,
		EscalationCount:    0,
		QRVerificationCode: "ETH-MUNI-TKT-2026-100234-VERIFIED",
		CreatedAt:          now.Add(-12 * time.Hour),
		UpdatedAt:          now,
	}

	// Case 2: Breached Case (ready for SLA automated escalation)
	case2 := domain.Case{
		ID:                 uuid.New(),
		TicketNumber:       "TKT-2026-100582",
		CitizenID:          citizen2.ID,
		ServiceTypeID:      serviceTypes[1].ID,
		CurrentStructureID: woreda01.ID,
		Title:              "Unresolved Commercial License Renewal Delay",
		Description:        "Application submitted over 3 days ago for retail pharmacy license renewal. Official did not sign off within mandated 24-hour SLA.",
		Status:             domain.StatusSubmitted,
		Priority:           domain.PriorityUrgent,
		SLADeadline:        now.Add(-6 * time.Hour), // BREACHED!
		IsEscalated:        false,
		EscalationCount:    0,
		QRVerificationCode: "ETH-MUNI-TKT-2026-100582-BREACH",
		CreatedAt:          now.Add(-30 * time.Hour),
		UpdatedAt:          now.Add(-30 * time.Hour),
	}

	// Case 3: Scheduled Wednesday Digital Hearing
	nextWednesday := getNextWeekday(now, time.Wednesday)
	case3 := domain.Case{
		ID:                 uuid.New(),
		TicketNumber:       "TKT-2026-100891",
		CitizenID:          citizen3.ID,
		ServiceTypeID:      serviceTypes[2].ID,
		CurrentStructureID: woreda02.ID,
		AssignedToUserID:   &woredaOfficer.ID,
		Title:              "Kebele House Tenancy Transfer Dispute",
		Description:        "Deceased parent's tenancy transfer denied without written justification. Requesting direct hearing with presiding officer.",
		Status:             domain.StatusScheduledForHearing,
		Priority:           domain.PriorityNormal,
		SLADeadline:        now.Add(48 * time.Hour),
		IsEscalated:        false,
		EscalationCount:    0,
		QRVerificationCode: "ETH-MUNI-TKT-2026-100891-HEARING",
		CreatedAt:          now.Add(-24 * time.Hour),
		UpdatedAt:          now,
	}

	db.Create(&case1)
	db.Create(&case2)
	db.Create(&case3)

	// Hearing slot for Case 3
	hearingSlot := domain.HearingSlot{
		ID:           uuid.New(),
		CaseID:       case3.ID,
		OfficialID:   woredaOfficer.ID,
		HearingDate:  nextWednesday,
		StartTime:    "10:00",
		EndTime:      "10:30",
		HearingType:  domain.HearingTypeVirtualCall,
		MeetingLink:  "https://meet.jit.si/eth-muni-hearing-TKT-2026-100891",
		Status:       domain.HearingStatusScheduled,
		HearingNotes: "Citizen requested virtual call to present original tenancy contract from 1998 E.C.",
		CreatedAt:    now,
	}
	db.Create(&hearingSlot)

	// Seed Initial Audit Logs
	initLog1, _ := json.Marshal(map[string]interface{}{"status": "SUBMITTED", "woreda": "Kirkos Woreda 01"})
	db.Create(&domain.CaseAuditLog{
		CaseID:    case1.ID,
		Action:    "CASE_INTAKE_SUBMITTED",
		NewValue:  string(initLog1),
		Notes:     "Citizen Tewodros Assefa submitted grievance via One-Stop Portal",
		CreatedAt: case1.CreatedAt,
	})
	db.Create(&domain.CaseAuditLog{
		CaseID:    case2.ID,
		Action:    "CASE_INTAKE_SUBMITTED",
		NewValue:  string(initLog1),
		Notes:     "Citizen Genet Wolde submitted grievance via One-Stop Portal",
		CreatedAt: case2.CreatedAt,
	})

	log.Println("[DB Seeder] Database seeded with 3 municipal tiers, 5 service types, 5 staff accounts, and 3 active/breached cases.")
	return nil
}

func getNextWeekday(from time.Time, target time.Weekday) time.Time {
	days := (int(target) - int(from.Weekday()) + 7) % 7
	if days == 0 {
		days = 7
	}
	next := from.AddDate(0, 0, days)
	return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, time.UTC)
}
