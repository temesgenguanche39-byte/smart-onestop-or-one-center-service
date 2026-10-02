package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/smart-onestop/platform/internal/application/dto"
	"github.com/smart-onestop/platform/internal/domain"
	"github.com/smart-onestop/platform/internal/domain/workcalendar"
	"github.com/smart-onestop/platform/internal/infrastructure/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	dbName := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open sqlite memory db: %v", err)
	}

	err = db.AutoMigrate(
		&domain.AdministrativeStructure{},
		&domain.User{},
		&domain.Citizen{},
		&domain.ServiceType{},
		&domain.Case{},
		&domain.CaseAttachment{},
		&domain.HearingSlot{},
		&domain.CaseAuditLog{},
		&domain.EscalationPolicy{},
		&domain.Holiday{},
	)
	if err != nil {
		t.Fatalf("Failed to auto migrate: %v", err)
	}

	return db
}

// TestSLASweepMustNotEscalateDuringHoliday tests that an SLA sweep never escalates tickets during a holiday or non-working day.
func TestSLASweepMustNotEscalateDuringHoliday(t *testing.T) {
	db := setupTestDB(t)

	caseRepo := repository.NewGormCaseRepository(db)
	structureRepo := repository.NewGormStructureRepository(db)
	auditRepo := repository.NewGormAuditRepository(db)
	holidayRepo := repository.NewGormHolidayRepository(db)

	calendarService := NewCalendarService(holidayRepo, auditRepo)
	slaService := NewSLAService(caseRepo, structureRepo, auditRepo)

	// Create hierarchy: City -> Sub-City -> Woreda
	city := domain.AdministrativeStructure{Name: "City Admin", Level: domain.AdminLevelCity, Code: "ETH-CITY"}
	db.Create(&city)
	subcity := domain.AdministrativeStructure{Name: "Kirkos Subcity", Level: domain.AdminLevelSubCity, Code: "ETH-KIR", ParentID: &city.ID}
	db.Create(&subcity)
	woreda := domain.AdministrativeStructure{Name: "Woreda 01", Level: domain.AdminLevelWoreda, Code: "ETH-W01", ParentID: &subcity.ID}
	db.Create(&woreda)

	citizen := domain.Citizen{
		ID:          uuid.New(),
		FullName:    "Abebe Bikila",
		PhoneNumber: "+251911223344",
	}
	db.Create(&citizen)

	st := domain.ServiceType{Name: "Land Admin", Code: "LAND", BaseSLAHours: 24}
	db.Create(&st)

	now := time.Now().UTC()

	// Create breached case
	caseEntity := domain.Case{
		ID:                 uuid.New(),
		TicketNumber:       "TKT-TEST-BREACH-01",
		CitizenID:          citizen.ID,
		ServiceTypeID:      st.ID,
		CurrentStructureID: woreda.ID,
		Title:              "Urgent Survey Dispute",
		Description:        "Boundary dispute delayed beyond SLA deadline",
		Status:             domain.StatusSubmitted,
		Priority:           domain.PriorityUrgent,
		SLADeadline:        now.Add(-2 * time.Hour), // Breached
		IsEscalated:        false,
		EscalationCount:    0,
		CreatedAt:          now.Add(-26 * time.Hour),
		UpdatedAt:          now.Add(-26 * time.Hour),
	}
	db.Create(&caseEntity)

	// Add a holiday for today in Africa/Addis_Ababa
	todayDate := now.In(workcalendar.AddisAbabaLocation())
	holiday := domain.Holiday{
		Date:   todayDate,
		NameEN: "National Test Holiday",
		NameAM: "የሙከራ ብሔራዊ በዓል",
		Type:   domain.HolidayTypeFixed,
		Active: true,
		Source: "Test Suite",
		Year:   todayDate.Year(),
	}
	db.Create(&holiday)
	calendarService.RefreshHolidayCache(context.Background())

	// Verify domain engine sees today as non-working day
	if workcalendar.IsWorkingDay(now, "") {
		t.Fatalf("Expected today to be non-working day because of holiday, got working day")
	}

	// 1. Run SLA sweep on holiday: MUST NOT ESCALATE!
	escalated, err := slaService.SweepAndEscalate(context.Background())
	if err != nil {
		t.Fatalf("Sweep failed: %v", err)
	}
	if escalated != 0 {
		t.Fatalf("Expected 0 cases escalated during holiday, got %d", escalated)
	}

	// Verify ticket remains at Woreda level
	var checkCase domain.Case
	db.First(&checkCase, caseEntity.ID)
	if checkCase.Status != domain.StatusSubmitted || checkCase.CurrentStructureID != woreda.ID || checkCase.IsEscalated {
		t.Fatalf("Case should NOT have escalated during holiday. Got status=%s, structID=%d, isEscalated=%v",
			checkCase.Status, checkCase.CurrentStructureID, checkCase.IsEscalated)
	}

	// 2. Deactivate the holiday
	holiday.Active = false
	db.Save(&holiday)
	calendarService.RefreshHolidayCache(context.Background())

	// If today is a weekend, temporarily override for testing
	cal := workcalendar.NewCalendar(workcalendar.Config{
		TimeZone: workcalendar.DefaultTimeZone,
		WorkDays: []time.Weekday{
			time.Sunday, time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday,
		},
		StartHour: 8, StartMinute: 30, EndHour: 17, EndMinute: 0,
	}, nil)
	_ = cal
}

// TestHearingValidationRejectsHoliday tests that a Wednesday or Friday falling on a public holiday
// is rejected with a suggestion for the next valid slot.
func TestHearingValidationRejectsHoliday(t *testing.T) {
	db := setupTestDB(t)

	caseRepo := repository.NewGormCaseRepository(db)
	auditRepo := repository.NewGormAuditRepository(db)
	hearingRepo := repository.NewGormHearingRepository(db)
	holidayRepo := repository.NewGormHolidayRepository(db)
	userRepo := repository.NewGormUserRepository(db)

	calendarService := NewCalendarService(holidayRepo, auditRepo)
	hearingService := NewHearingService(hearingRepo, caseRepo, auditRepo, userRepo)

	city := domain.AdministrativeStructure{Name: "City Admin", Level: domain.AdminLevelCity, Code: "ETH-CITY"}
	db.Create(&city)
	citizen := domain.Citizen{ID: uuid.New(), FullName: "Mulugeta", PhoneNumber: "+251911998877"}
	db.Create(&citizen)
	st := domain.ServiceType{Name: "Housing", Code: "HOUSE", BaseSLAHours: 48}
	db.Create(&st)

	caseEntity := domain.Case{
		ID:                 uuid.New(),
		TicketNumber:       "TKT-HEARING-TEST",
		CitizenID:          citizen.ID,
		ServiceTypeID:      st.ID,
		CurrentStructureID: city.ID,
		Title:              "Dispute",
		Description:        "Hearing required",
		Status:             domain.StatusSubmitted,
		Priority:           domain.PriorityNormal,
		SLADeadline:        time.Now().Add(48 * time.Hour),
		QRVerificationCode: "ETH-TEST-QR-HEARING",
	}
	if err := db.Create(&caseEntity).Error; err != nil {
		t.Fatalf("Failed to create test case: %v", err)
	}

	// Friday 2026-10-09
	fridayHolidayStr := "2026-10-09"
	loc := workcalendar.AddisAbabaLocation()
	friDate, _ := time.ParseInLocation("2006-01-02", fridayHolidayStr, loc)

	if friDate.Weekday() != time.Friday {
		t.Fatalf("2026-10-09 must be Friday")
	}

	// Register 2026-10-09 as a public holiday
	h := domain.Holiday{
		Date:   friDate,
		NameEN: "Test Friday Holiday",
		NameAM: "የሙከራ ዓርብ በዓል",
		Type:   domain.HolidayTypeFixed,
		Active: true,
		Source: "Council of Ministers",
		Year:   2026,
	}
	db.Create(&h)
	calendarService.RefreshHolidayCache(context.Background())

	officialID := uuid.New()
	req := dto.ScheduleHearingRequest{
		HearingDate: fridayHolidayStr,
		StartTime:   "10:00",
		EndTime:     "10:30",
		HearingType: domain.HearingTypeVirtualCall,
	}

	// Attempt scheduling on holiday Friday
	_, err := hearingService.ScheduleSlot(context.Background(), caseEntity.ID, officialID, req)
	if err == nil {
		t.Fatalf("Expected error when scheduling hearing on a public holiday Friday, got nil")
	}

	var holErr *HearingHolidayError
	if !errors.As(err, &holErr) {
		t.Fatalf("Expected HearingHolidayError, got %v", err)
	}

	// Suggested slot should be next Wednesday: 2026-10-14
	expectedNextSlot := "2026-10-14"
	actualNextSlot := holErr.SuggestedSlot.Format("2006-01-02")
	if actualNextSlot != expectedNextSlot {
		t.Errorf("Expected suggested slot %s, got %s", expectedNextSlot, actualNextSlot)
	}
}

// TestCalendarServiceWorkingDaysTableDriven tests working days counting between Gregorian and Ethiopian inputs
func TestCalendarServiceWorkingDaysTableDriven(t *testing.T) {
	db := setupTestDB(t)
	holidayRepo := repository.NewGormHolidayRepository(db)
	auditRepo := repository.NewGormAuditRepository(db)
	calService := NewCalendarService(holidayRepo, auditRepo)

	// Insert a holiday on Monday 2026-10-05
	loc := workcalendar.AddisAbabaLocation()
	hDate := time.Date(2026, 10, 5, 0, 0, 0, 0, loc)
	db.Create(&domain.Holiday{
		Date:   hDate,
		NameEN: "Test Monday Holiday",
		NameAM: "የሙከራ ሰኞ በዓል",
		Type:   domain.HolidayTypeFixed,
		Active: true,
		Year:   2026,
	})
	calService.RefreshHolidayCache(context.Background())

	testCases := []struct {
		name                string
		from                string
		to                  string
		calendar            string
		expectedHolidays    int
		minWorkingDays      float64
		maxWorkingDays      float64
	}{
		{
			name:             "Friday to following Friday with Monday holiday",
			from:             "2026-10-02", // Friday
			to:               "2026-10-09", // Friday
			calendar:         "gregorian",
			expectedHolidays: 1, // Oct 5
			minWorkingDays:   4.5,
			maxWorkingDays:   5.5,
		},
		{
			name:             "Ethiopian calendar input (Meskerem 22 to Meskerem 29, 2019)",
			from:             "2019-01-22",
			to:               "2019-01-29",
			calendar:         "ethiopian",
			expectedHolidays: 1,
			minWorkingDays:   4.5,
			maxWorkingDays:   5.5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := calService.GetWorkingDays(context.Background(), dto.WorkingDaysRequest{
				From:     tc.from,
				To:       tc.to,
				Calendar: tc.calendar,
			})
			if err != nil {
				t.Fatalf("GetWorkingDays failed: %v", err)
			}

			if resp.HolidaysEncountered != tc.expectedHolidays {
				t.Errorf("Expected %d holidays, got %d", tc.expectedHolidays, resp.HolidaysEncountered)
			}
			if resp.WorkingDays < tc.minWorkingDays || resp.WorkingDays > tc.maxWorkingDays {
				t.Errorf("Working days %f out of expected range [%f, %f]", resp.WorkingDays, tc.minWorkingDays, tc.maxWorkingDays)
			}
			if resp.FromEthiopian == "" || resp.FromEthiopianAM == "" {
				t.Errorf("Missing Ethiopian date formatting in response")
			}
		})
	}
}
