package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/smart-onestop/platform/internal/application/dto"
	"github.com/smart-onestop/platform/internal/domain"
	"github.com/smart-onestop/platform/internal/domain/workcalendar"
)

type CalendarService struct {
	holidayRepo domain.HolidayRepository
	auditRepo   domain.AuditRepository
	holidayMu   sync.RWMutex
	holidayMap  map[string]bool // "YYYY-MM-DD" -> true
}

func NewCalendarService(
	holidayRepo domain.HolidayRepository,
	auditRepo domain.AuditRepository,
) *CalendarService {
	s := &CalendarService{
		holidayRepo: holidayRepo,
		auditRepo:   auditRepo,
		holidayMap:  make(map[string]bool),
	}

	// Prime in-memory holiday cache and register domain checker
	s.RefreshHolidayCache(context.Background())

	workcalendar.SetDefaultHolidayChecker(func(date time.Time, structureID string) bool {
		return s.IsHoliday(date, structureID)
	})

	return s
}

// RefreshHolidayCache loads active holidays into memory for high-performance pure lookups
func (s *CalendarService) RefreshHolidayCache(ctx context.Context) {
	s.holidayMu.Lock()
	defer s.holidayMu.Unlock()

	holidays, err := s.holidayRepo.ListAll(ctx)
	if err != nil {
		return
	}

	newMap := make(map[string]bool)
	for _, h := range holidays {
		if h.Active {
			dateKey := h.Date.Format("2006-01-02")
			newMap[dateKey] = true
		}
	}
	s.holidayMap = newMap
}

// IsHoliday checks if a date is a registered holiday
func (s *CalendarService) IsHoliday(date time.Time, structureID string) bool {
	s.holidayMu.RLock()
	defer s.holidayMu.RUnlock()

	dateKey := date.In(workcalendar.AddisAbabaLocation()).Format("2006-01-02")
	return s.holidayMap[dateKey]
}

// ParseDateString parses a date string according to calendar type ("ethiopian" or "gregorian")
func (s *CalendarService) ParseDateString(dateStr string, calType string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	parts := strings.Split(dateStr, "-")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("invalid date format %q: expected YYYY-MM-DD", dateStr)
	}

	y, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid year in %q", dateStr)
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid month in %q", dateStr)
	}
	d, err := strconv.Atoi(parts[2])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid day in %q", dateStr)
	}

	if strings.ToLower(calType) == "ethiopian" {
		if m < 1 || m > 13 {
			return time.Time{}, fmt.Errorf("invalid Ethiopian month %d (must be 1 to 13)", m)
		}
		maxDays := workcalendar.DaysInEthiopianMonth(y, m)
		if d < 1 || d > maxDays {
			return time.Time{}, fmt.Errorf("invalid Ethiopian day %d for month %d year %d (max %d)", d, m, y, maxDays)
		}
		return workcalendar.ToGregorianDate(workcalendar.EthiopianDate{Year: y, Month: m, Day: d}), nil
	}

	// Gregorian
	loc := workcalendar.AddisAbabaLocation()
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, loc), nil
}

// GetWorkingDays counts working days between from and to, detailing holidays and weekends
func (s *CalendarService) GetWorkingDays(ctx context.Context, req dto.WorkingDaysRequest) (*dto.WorkingDaysResponse, error) {
	fromDate, err := s.ParseDateString(req.From, req.Calendar)
	if err != nil {
		return nil, err
	}
	toDate, err := s.ParseDateString(req.To, req.Calendar)
	if err != nil {
		return nil, err
	}

	loc := workcalendar.AddisAbabaLocation()
	normFrom := time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), 0, 0, 0, 0, loc)
	normTo := time.Date(toDate.Year(), toDate.Month(), toDate.Day(), 0, 0, 0, 0, loc)

	if normFrom.After(normTo) {
		normFrom, normTo = normTo, normFrom
	}

	// Calculate working days through domain engine
	startOffice := time.Date(normFrom.Year(), normFrom.Month(), normFrom.Day(), 8, 30, 0, 0, loc)
	endOffice := time.Date(normTo.Year(), normTo.Month(), normTo.Day(), 17, 0, 0, 0, loc)
	workingDays := workcalendar.CountWorkingDays(startOffice, endOffice)

	totalDays := 0
	holidaysCount := 0
	weekendDays := 0

	curr := normFrom
	for !curr.After(normTo) {
		totalDays++
		wd := curr.Weekday()
		if wd == time.Sunday || wd == time.Saturday {
			weekendDays++
		} else if s.IsHoliday(curr, "") {
			holidaysCount++
		}
		curr = curr.AddDate(0, 0, 1)
	}

	fromEth := workcalendar.ToEthiopianDate(normFrom)
	toEth := workcalendar.ToEthiopianDate(normTo)

	return &dto.WorkingDaysResponse{
		FromGregorian:       normFrom.Format("2006-01-02"),
		ToGregorian:         normTo.Format("2006-01-02"),
		FromEthiopian:       workcalendar.FormatEthiopian(fromEth, "en"),
		ToEthiopian:         workcalendar.FormatEthiopian(toEth, "en"),
		FromEthiopianAM:     workcalendar.FormatEthiopian(fromEth, "am"),
		ToEthiopianAM:       workcalendar.FormatEthiopian(toEth, "am"),
		WorkingDays:         workingDays,
		TotalCalendarDays:   totalDays,
		HolidaysEncountered: holidaysCount,
		WeekendDaysSkipped:  weekendDays,
	}, nil
}

// GetHolidays returns all public holidays for a given year (or all if year <= 0)
func (s *CalendarService) GetHolidays(ctx context.Context, year int) ([]dto.HolidayResponse, error) {
	holidays, err := s.holidayRepo.ListByYear(ctx, year)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.HolidayResponse, 0, len(holidays))
	for _, h := range holidays {
		ethDate := workcalendar.ToEthiopianDate(h.Date)
		responses = append(responses, dto.HolidayResponse{
			ID:              h.ID,
			Date:            h.Date.Format("2006-01-02"),
			EthiopianDate:   workcalendar.FormatEthiopian(ethDate, "en"),
			EthiopianDateAM: workcalendar.FormatEthiopian(ethDate, "am"),
			NameEN:          h.NameEN,
			NameAM:          h.NameAM,
			Type:            h.Type,
			Active:          h.Active,
			Source:          h.Source,
			Year:            h.Year,
			StructureID:     h.StructureID,
		})
	}
	return responses, nil
}

// GetNextHearingSlots returns the next N Wednesday and Friday hearing dates
func (s *CalendarService) GetNextHearingSlots(ctx context.Context, fromStr string, count int) (*dto.NextHearingSlotsResponse, error) {
	loc := workcalendar.AddisAbabaLocation()
	var fromDate time.Time
	if strings.TrimSpace(fromStr) != "" {
		parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(fromStr), loc)
		if err == nil {
			fromDate = parsed
		} else {
			fromDate = time.Now().In(loc)
		}
	} else {
		fromDate = time.Now().In(loc)
	}

	if count <= 0 {
		count = 5
	}
	if count > 30 {
		count = 30
	}

	slots := workcalendar.NextHearingSlots(fromDate, count)
	slotResponses := make([]dto.HearingSlotOption, 0, len(slots))

	for _, slot := range slots {
		ethDate := workcalendar.ToEthiopianDate(slot)
		slotWd := slot.Weekday()
		wdEN := "Wednesday"
		wdAM := "ረቡዕ"
		if slotWd == time.Friday {
			wdEN = "Friday"
			wdAM = "ዓርብ"
		}

		slotResponses = append(slotResponses, dto.HearingSlotOption{
			Date:            slot.Format("2006-01-02"),
			EthiopianDate:   workcalendar.FormatEthiopian(ethDate, "en"),
			EthiopianDateAM: workcalendar.FormatEthiopian(ethDate, "am"),
			DayOfWeek:       wdEN,
			DayOfWeekAM:     wdAM,
			IsWorkingDay:    true,
		})
	}

	return &dto.NextHearingSlotsResponse{
		From:  fromDate.Format("2006-01-02"),
		Count: len(slotResponses),
		Slots: slotResponses,
	}, nil
}

// CreateHoliday allows SUPER_ADMIN to add a holiday
func (s *CalendarService) CreateHoliday(ctx context.Context, req dto.CreateHolidayRequest) (*dto.HolidayResponse, error) {
	loc := workcalendar.AddisAbabaLocation()
	parsedDate, err := time.ParseInLocation("2006-01-02", req.Date, loc)
	if err != nil {
		return nil, domain.ErrInvalidDateFormat
	}

	year := req.Year
	if year == 0 {
		year = parsedDate.Year()
	}

	active := true
	if req.Active != nil {
		active = *req.Active
	}

	holiday := &domain.Holiday{
		Date:        parsedDate,
		NameEN:      req.NameEN,
		NameAM:      req.NameAM,
		Type:        req.Type,
		Active:      active,
		Source:      req.Source,
		Year:        year,
		StructureID: req.StructureID,
	}

	if err := s.holidayRepo.Create(ctx, holiday); err != nil {
		return nil, err
	}

	s.RefreshHolidayCache(ctx)

	ethDate := workcalendar.ToEthiopianDate(holiday.Date)
	return &dto.HolidayResponse{
		ID:              holiday.ID,
		Date:            holiday.Date.Format("2006-01-02"),
		EthiopianDate:   workcalendar.FormatEthiopian(ethDate, "en"),
		EthiopianDateAM: workcalendar.FormatEthiopian(ethDate, "am"),
		NameEN:          holiday.NameEN,
		NameAM:          holiday.NameAM,
		Type:            holiday.Type,
		Active:          holiday.Active,
		Source:          holiday.Source,
		Year:            holiday.Year,
		StructureID:     holiday.StructureID,
	}, nil
}

// UpdateHoliday allows SUPER_ADMIN to modify a holiday
func (s *CalendarService) UpdateHoliday(ctx context.Context, id uint, req dto.UpdateHolidayRequest) (*dto.HolidayResponse, error) {
	h, err := s.holidayRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	loc := workcalendar.AddisAbabaLocation()
	if strings.TrimSpace(req.Date) != "" {
		parsedDate, err := time.ParseInLocation("2006-01-02", req.Date, loc)
		if err != nil {
			return nil, domain.ErrInvalidDateFormat
		}
		h.Date = parsedDate
		if req.Year == 0 {
			h.Year = parsedDate.Year()
		}
	}

	if strings.TrimSpace(req.NameEN) != "" {
		h.NameEN = req.NameEN
	}
	if strings.TrimSpace(req.NameAM) != "" {
		h.NameAM = req.NameAM
	}
	if req.Type != "" {
		h.Type = req.Type
	}
	if req.Active != nil {
		h.Active = *req.Active
	}
	if req.Source != "" {
		h.Source = req.Source
	}
	if req.Year > 0 {
		h.Year = req.Year
	}
	if req.StructureID != nil {
		h.StructureID = req.StructureID
	}

	if err := s.holidayRepo.Update(ctx, h); err != nil {
		return nil, err
	}

	s.RefreshHolidayCache(ctx)

	ethDate := workcalendar.ToEthiopianDate(h.Date)
	return &dto.HolidayResponse{
		ID:              h.ID,
		Date:            h.Date.Format("2006-01-02"),
		EthiopianDate:   workcalendar.FormatEthiopian(ethDate, "en"),
		EthiopianDateAM: workcalendar.FormatEthiopian(ethDate, "am"),
		NameEN:          h.NameEN,
		NameAM:          h.NameAM,
		Type:            h.Type,
		Active:          h.Active,
		Source:          h.Source,
		Year:            h.Year,
		StructureID:     h.StructureID,
	}, nil
}

// DeleteHoliday deletes a holiday by ID
func (s *CalendarService) DeleteHoliday(ctx context.Context, id uint) error {
	if err := s.holidayRepo.Delete(ctx, id); err != nil {
		return err
	}
	s.RefreshHolidayCache(ctx)
	return nil
}
