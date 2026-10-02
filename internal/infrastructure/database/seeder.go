package database

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/smart-onestop/platform/internal/domain"
	"github.com/smart-onestop/platform/internal/domain/workcalendar"
	"github.com/smart-onestop/platform/internal/infrastructure/security"
	"gorm.io/gorm"
)

// SeedDatabase inserts initial municipal hierarchy, service types, users, and demonstration cases
func SeedDatabase(db *gorm.DB) error {
	var count int64
	db.Model(&domain.AdministrativeStructure{}).Count(&count)
	if count > 0 {
		log.Println("[DB Seeder] Database already populated with seed data.")
		_ = SeedHolidays(db)
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

	type subCityDef struct {
		Name  string
		Code  string
		Count int
	}

	subCitiesList := []subCityDef{
		{"አዲስ ከተማ (Addis Ketema)", "ETH-AA-ADK", 14},
		{"አራዳ (Arada)", "ETH-AA-ARD", 10},
		{"ቂርቆስ (Kirkos)", "ETH-AA-KIR", 11},
		{"ልደታ (Lideta)", "ETH-AA-LID", 10},
		{"የካ (Yeka)", "ETH-AA-YEK", 14},
		{"ቦሌ (Bole)", "ETH-AA-BOL", 15},
		{"ጉለሌ (Gullele)", "ETH-AA-GUL", 11},
		{"ኮልፌ ቀራኒዮ (Kolfe Keranio)", "ETH-AA-KOL", 15},
		{"ንፋስ ስልክ ላፍቶ (Nifas Silk Lafto)", "ETH-AA-NSL", 15},
		{"አካቂ ቃሊቲ (Akaki Kality)", "ETH-AA-AKK", 13},
		{"ለሚ ኩራ (Lemi Kura)", "ETH-AA-LMK", 10},
	}

	var kirkosSubCityID, kirkosWoreda01ID uint
	for _, sc := range subCitiesList {
		subCityRecord := domain.AdministrativeStructure{
			Name:      sc.Name,
			Level:     domain.AdminLevelSubCity,
			ParentID:  &city.ID,
			Code:      sc.Code,
			CreatedAt: time.Now().UTC(),
		}
		db.Create(&subCityRecord)
		if sc.Code == "ETH-AA-KIR" {
			kirkosSubCityID = subCityRecord.ID
		}

		for w := 1; w <= sc.Count; w++ {
			wCode := fmt.Sprintf("%s-W%02d", sc.Code, w)
			wName := fmt.Sprintf("%s ወረዳ %02d", strings.Split(sc.Name, " ")[0], w)
			woredaRecord := domain.AdministrativeStructure{
				Name:      wName,
				Level:     domain.AdminLevelWoreda,
				ParentID:  &subCityRecord.ID,
				Code:      wCode,
				CreatedAt: time.Now().UTC(),
			}
			db.Create(&woredaRecord)
			if sc.Code == "ETH-AA-KIR" && w == 1 {
				kirkosWoreda01ID = woredaRecord.ID
			}
		}
	}

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
			StructureID:  &kirkosSubCityID,
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
			StructureID:  &kirkosWoreda01ID,
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
			StructureID:  &kirkosWoreda01ID,
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
		WoredaID:          &kirkosWoreda01ID,
		HouseNumber:       "142/B",
		PreferredLanguage: "am",
		CreatedAt:         time.Now().UTC(),
	}
	citizen2 := domain.Citizen{
		ID:                uuid.New(),
		FullName:          "Genet Wolde",
		PhoneNumber:       "+251922334455",
		NationalID:        "ETH-NAT-774411",
		WoredaID:          &kirkosWoreda01ID,
		HouseNumber:       "089",
		PreferredLanguage: "am",
		CreatedAt:         time.Now().UTC(),
	}
	citizen3 := domain.Citizen{
		ID:                uuid.New(),
		FullName:          "Dawit Kebede",
		PhoneNumber:       "+251933445566",
		NationalID:        "ETH-NAT-556677",
		WoredaID:          &kirkosWoreda01ID,
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
		CurrentStructureID: kirkosWoreda01ID,
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
		CurrentStructureID: kirkosWoreda01ID,
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
		CurrentStructureID: kirkosWoreda01ID,
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
	_ = SeedHolidays(db)
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

// SeedHolidays seeds the canonical Ethiopian public, national, Orthodox and Islamic holidays
func SeedHolidays(db *gorm.DB) error {
	var count int64
	db.Model(&domain.Holiday{}).Count(&count)
	if count > 0 {
		return nil
	}

	log.Println("[DB Seeder] Seeding Ethiopian canonical public holidays across 2024-2028...")
	loc := workcalendar.AddisAbabaLocation()

	// Islamic Moon-based holidays verified from Ethiopian Supreme Council of Islamic Affairs
	type islamicHolidayEntry struct {
		Date   string
		NameEN string
		NameAM string
		Year   int
	}

	islamicHolidays := []islamicHolidayEntry{
		// 2024
		{"2024-04-10", "Eid al-Fitr", "ዒድ አል-ፊጥር (የጾም ፍቺ በዓል)", 2024},
		{"2024-06-17", "Eid al-Adha (Arefa)", "ዒድ አል-አድሃ (አረፋ)", 2024},
		{"2024-09-16", "Mawlid (Birth of the Prophet)", "መውሊድ (የነቢዩ ሙሐመድ የልደት በዓል)", 2024},
		// 2025
		{"2025-03-31", "Eid al-Fitr", "ዒድ አል-ፊጥር (የጾም ፍቺ በዓል)", 2025},
		{"2025-06-07", "Eid al-Adha (Arefa)", "ዒድ አል-አድሃ (አረፋ)", 2025},
		{"2025-09-05", "Mawlid (Birth of the Prophet)", "መውሊድ (የነቢዩ ሙሐመድ የልደት በዓል)", 2025},
		// 2026
		{"2026-03-20", "Eid al-Fitr", "ዒድ አል-ፊጥር (የጾም ፍቺ በዓል)", 2026},
		{"2026-05-27", "Eid al-Adha (Arefa)", "ዒድ አል-አድሃ (አረፋ)", 2026},
		{"2026-08-25", "Mawlid (Birth of the Prophet)", "መውሊድ (የነቢዩ ሙሐመድ የልደት በዓል)", 2026},
		// 2027
		{"2027-03-10", "Eid al-Fitr", "ዒድ አል-ፊጥር (የጾም ፍቺ በዓል)", 2027},
		{"2027-05-17", "Eid al-Adha (Arefa)", "ዒድ አል-አድሃ (አረፋ)", 2027},
		{"2027-08-15", "Mawlid (Birth of the Prophet)", "መውሊድ (የነቢዩ ሙሐመድ የልደት በዓል)", 2027},
		// 2028
		{"2028-02-27", "Eid al-Fitr", "ዒድ አል-ፊጥር (የጾም ፍቺ በዓል)", 2028},
		{"2028-05-05", "Eid al-Adha (Arefa)", "ዒድ አል-አድሃ (አረፋ)", 2028},
		{"2028-08-04", "Mawlid (Birth of the Prophet)", "መውሊድ (የነቢዩ ሙሐመድ የልደት በዓል)", 2028},
	}

	const source = "Federal Democratic Republic of Ethiopia Public Holidays (Proclamation No. 16/1975)"

	for y := 2024; y <= 2028; y++ {
		// Ethiopian Fixed Holidays in early Gregorian year (Ethiopian Year: y - 8)
		ethEarlyYear := y - 8
		gennaDate := workcalendar.ToGregorianDate(workcalendar.EthiopianDate{Year: ethEarlyYear, Month: 4, Day: 29})
		timketDate := workcalendar.ToGregorianDate(workcalendar.EthiopianDate{Year: ethEarlyYear, Month: 5, Day: 11})
		adwaDate := workcalendar.ToGregorianDate(workcalendar.EthiopianDate{Year: ethEarlyYear, Month: 6, Day: 23})
		labourDate := time.Date(y, time.May, 1, 0, 0, 0, 0, loc)
		patriotsDate := workcalendar.ToGregorianDate(workcalendar.EthiopianDate{Year: ethEarlyYear, Month: 8, Day: 27})
		dergDate := workcalendar.ToGregorianDate(workcalendar.EthiopianDate{Year: ethEarlyYear, Month: 9, Day: 20})

		// Ethiopian Fixed Holidays in late Gregorian year (Ethiopian Year: y - 7)
		ethLateYear := y - 7
		enkutatashDate := workcalendar.ToGregorianDate(workcalendar.EthiopianDate{Year: ethLateYear, Month: 1, Day: 1})
		meskelDate := workcalendar.ToGregorianDate(workcalendar.EthiopianDate{Year: ethLateYear, Month: 1, Day: 17})

		// Orthodox Movable Holidays
		sikletDate := workcalendar.OrthodoxGoodFriday(y)
		fasikaDate := workcalendar.OrthodoxEaster(y)

		holidays := []domain.Holiday{
			{Date: gennaDate, NameEN: "Genna (Ethiopian Christmas)", NameAM: "ገና (የገና / የልደት በዓል)", Type: domain.HolidayTypeFixed, Active: true, Source: source, Year: y},
			{Date: timketDate, NameEN: "Timket (Epiphany)", NameAM: "ጥምቀት (የጥምቀት በዓል)", Type: domain.HolidayTypeFixed, Active: true, Source: source, Year: y},
			{Date: adwaDate, NameEN: "Adwa Victory Day", NameAM: "የአድዋ ድል በዓል", Type: domain.HolidayTypeFixed, Active: true, Source: source, Year: y},
			{Date: labourDate, NameEN: "International Labour Day", NameAM: "የሰራተኞች ቀን (ሜይ ዴይ)", Type: domain.HolidayTypeFixed, Active: true, Source: source, Year: y},
			{Date: patriotsDate, NameEN: "Patriots' Victory Day", NameAM: "የአርበኞች ቀን", Type: domain.HolidayTypeFixed, Active: true, Source: source, Year: y},
			{Date: dergDate, NameEN: "Derg Downfall Day", NameAM: "የደርግ ውድቀት ቀን (ግንቦት 20)", Type: domain.HolidayTypeFixed, Active: true, Source: source, Year: y},
			{Date: enkutatashDate, NameEN: "Enkutatash (Ethiopian New Year)", NameAM: "እንቁጣጣሽ (የአዲስ ዓመት በዓል)", Type: domain.HolidayTypeFixed, Active: true, Source: source, Year: y},
			{Date: meskelDate, NameEN: "Meskel (Finding of the True Cross)", NameAM: "መስቀል (የመስቀል ደመራ በዓል)", Type: domain.HolidayTypeFixed, Active: true, Source: source, Year: y},
			{Date: sikletDate, NameEN: "Good Friday (Siklet)", NameAM: "ስቅለት (የስቅለት በዓል)", Type: domain.HolidayTypeMovable, Active: true, Source: source, Year: y},
			{Date: fasikaDate, NameEN: "Easter (Fasika)", NameAM: "ፋሲካ (የትንሣኤ በዓል)", Type: domain.HolidayTypeMovable, Active: true, Source: source, Year: y},
		}

		for _, h := range holidays {
			db.Create(&h)
		}
	}

	// Seed Islamic moon-based holidays
	for _, ih := range islamicHolidays {
		d, _ := time.ParseInLocation("2006-01-02", ih.Date, loc)
		h := domain.Holiday{
			Date:   d,
			NameEN: ih.NameEN,
			NameAM: ih.NameAM,
			Type:   domain.HolidayTypeIslamic,
			Active: true,
			Source: "Ethiopian Supreme Council of Islamic Affairs (Moon-Sighting Observation)",
			Year:   ih.Year,
		}
		db.Create(&h)
	}

	log.Println("[DB Seeder] Canonical Ethiopian holidays seeded successfully.")
	return nil
}
