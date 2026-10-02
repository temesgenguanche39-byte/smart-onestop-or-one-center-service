package workcalendar

import (
	"testing"
	"time"
)

// TestEthiopianCalendarRoundTripAcross100Years tests round-trip conversion for every day from 1920 to 2030 (over 100 years).
func TestEthiopianCalendarRoundTripAcross100Years(t *testing.T) {
	loc := AddisAbabaLocation()
	startDate := time.Date(1920, 1, 1, 0, 0, 0, 0, loc)
	endDate := time.Date(2030, 1, 1, 0, 0, 0, 0, loc)

	curr := startDate
	count := 0
	for curr.Before(endDate) {
		ed := ToEthiopianDate(curr)

		// Validate month range
		if ed.Month < 1 || ed.Month > 13 {
			t.Fatalf("Invalid month %d on %v", ed.Month, curr)
		}
		// Validate day range
		maxDays := DaysInEthiopianMonth(ed.Year, ed.Month)
		if ed.Day < 1 || ed.Day > maxDays {
			t.Fatalf("Invalid day %d in month %d year %d (max %d) on %v", ed.Day, ed.Month, ed.Year, maxDays, curr)
		}

		// Round trip back to Gregorian
		greg := ToGregorianDate(ed)
		if greg.Year() != curr.Year() || greg.Month() != curr.Month() || greg.Day() != curr.Day() {
			t.Fatalf("Mismatch on %v: got Gregorian %v, Ethiopian was %+v", curr, greg, ed)
		}

		curr = curr.AddDate(0, 0, 1)
		count++
	}

	if count < 365*100 {
		t.Fatalf("Expected over 36,500 round trips, got %d", count)
	}
}

// TestPagumeLeapYears tests Pagume has 6 days when year % 4 == 3, and 5 days otherwise.
func TestPagumeLeapYears(t *testing.T) {
	testCases := []struct {
		ethYear      int
		isLeap       bool
		expectedDays int
	}{
		{2011, (2011 % 4) == 3, 6}, // 2011 % 4 = 3 -> Leap year (Pagume 6)
		{2012, false, 5},
		{2013, false, 5},
		{2014, false, 5},
		{2015, (2015 % 4) == 3, 6}, // 2015 % 4 = 3 -> Leap year (Pagume 6)
		{2016, false, 5},
		{2017, false, 5},
		{2018, false, 5},
		{2019, (2019 % 4) == 3, 6}, // 2019 % 4 = 3 -> Leap year (Pagume 6)
	}

	for _, tc := range testCases {
		isLeap := IsEthiopianLeapYear(tc.ethYear)
		if isLeap != tc.isLeap {
			t.Errorf("Year %d leap expectation mismatch: got %v, want %v", tc.ethYear, isLeap, tc.isLeap)
		}
		days := DaysInEthiopianMonth(tc.ethYear, 13)
		if days != tc.expectedDays {
			t.Errorf("Year %d Pagume days mismatch: got %d, want %d", tc.ethYear, days, tc.expectedDays)
		}
	}
}

// TestFixedHolidayShift tests that in Gregorian calendar, Ethiopian fixed holidays
// shift by +1 day following an Ethiopian leap year (when Pagume had 6 days).
func TestFixedHolidayShift(t *testing.T) {
	// Ethiopian 2015 was a leap year (Pagume had 6 days, Sep 6 to Sep 11, 2023).
	// Therefore, Meskerem 1, 2016 fell on Sep 12, 2023 (shifted by +1 day from normal Sep 11).
	// Ethiopian 2016 was not a leap year (Pagume had 5 days, Sep 6 to Sep 10, 2024).
	// Therefore, Meskerem 1, 2017 fell on Sep 11, 2024 (normal Sep 11).

	m1_2016 := ToGregorianDate(EthiopianDate{Year: 2016, Month: 1, Day: 1})
	if m1_2016.Year() != 2023 || m1_2016.Month() != time.September || m1_2016.Day() != 12 {
		t.Errorf("Expected Meskerem 1, 2016 to be 2023-09-12 (shifted by 1 day), got %v", m1_2016.Format("2006-01-02"))
	}

	m1_2017 := ToGregorianDate(EthiopianDate{Year: 2017, Month: 1, Day: 1})
	if m1_2017.Year() != 2024 || m1_2017.Month() != time.September || m1_2017.Day() != 11 {
		t.Errorf("Expected Meskerem 1, 2017 to be 2024-09-11 (normal), got %v", m1_2017.Format("2006-01-02"))
	}

	// Meskel (Meskerem 17):
	// In 2016 (after leap year): Sep 28, 2023
	// In 2017 (normal): Sep 27, 2024
	meskel2016 := ToGregorianDate(EthiopianDate{Year: 2016, Month: 1, Day: 17})
	if meskel2016.Day() != 28 || meskel2016.Month() != time.September {
		t.Errorf("Expected Meskel 2016 to be Sep 28, got %v", meskel2016.Format("2006-01-02"))
	}

	meskel2017 := ToGregorianDate(EthiopianDate{Year: 2017, Month: 1, Day: 17})
	if meskel2017.Day() != 27 || meskel2017.Month() != time.September {
		t.Errorf("Expected Meskel 2017 to be Sep 27, got %v", meskel2017.Format("2006-01-02"))
	}
}

// TestOrthodoxEasterAndGoodFriday2025Through2035 verifies the Orthodox computus against official ecclesiastical tables.
func TestOrthodoxEasterAndGoodFriday2025Through2035(t *testing.T) {
	expectedEaster := map[int]string{
		2025: "2025-04-20",
		2026: "2026-04-12",
		2027: "2027-05-02",
		2028: "2028-04-16",
		2029: "2029-04-08",
		2030: "2030-04-28",
		2031: "2031-04-13",
		2032: "2032-05-02",
		2033: "2033-04-24",
		2034: "2034-04-09",
		2035: "2035-04-29",
	}

	for year, expectedStr := range expectedEaster {
		easter := OrthodoxEaster(year)
		actualStr := easter.Format("2006-01-02")
		if actualStr != expectedStr {
			t.Errorf("Orthodox Easter for %d: got %s, want %s", year, actualStr, expectedStr)
		}

		// Good Friday (Siklet) must be Friday before Easter Sunday (2 days earlier)
		siklet := OrthodoxGoodFriday(year)
		if siklet.Weekday() != time.Friday {
			t.Errorf("Good Friday for %d must be Friday, got %v", year, siklet.Weekday())
		}
		if siklet.AddDate(0, 0, 2).Format("2006-01-02") != actualStr {
			t.Errorf("Good Friday for %d must be 2 days before Easter: got %v vs Easter %v", year, siklet, easter)
		}
	}
}

// TestTimezoneMidnightEdgeCase tests that timezone normalization to Africa/Addis_Ababa
// correctly handles UTC timestamps near midnight.
func TestTimezoneMidnightEdgeCase(t *testing.T) {
	loc := AddisAbabaLocation()

	// 2026-10-04 22:30 UTC is 2026-10-05 01:30 in Africa/Addis_Ababa (UTC+3)
	utcTime := time.Date(2026, 10, 4, 22, 30, 0, 0, time.UTC)
	cal := NewCalendar(DefaultConfig(), nil)
	norm := cal.NormalizeDate(utcTime)

	if norm.Day() != 5 || norm.Month() != time.October || norm.Year() != 2026 {
		t.Fatalf("Expected normalized date to be Oct 5 in Addis Ababa, got %v", norm)
	}

	if norm.Location().String() != loc.String() {
		t.Fatalf("Expected timezone %v, got %v", loc, norm.Location())
	}
}

// TestHolidayOnFridayAndNextHearingDay tests when a Friday is a public holiday,
// IsWorkingDay returns false, and NextHearingDay returns the following Wednesday.
func TestHolidayOnFridayAndNextHearingDay(t *testing.T) {
	loc := AddisAbabaLocation()

	// 2026-10-09 is a Friday
	friday := time.Date(2026, 10, 9, 10, 0, 0, 0, loc)
	if friday.Weekday() != time.Friday {
		t.Fatalf("2026-10-09 should be Friday")
	}

	// Mock holiday on 2026-10-09
	holidayChecker := func(date time.Time, structureID string) bool {
		return date.Year() == 2026 && date.Month() == time.October && date.Day() == 9
	}

	cal := NewCalendar(DefaultConfig(), holidayChecker)

	if cal.IsWorkingDay(friday, "") {
		t.Errorf("Friday 2026-10-09 is a holiday, should not be a working day")
	}

	// Next hearing day from Thursday 2026-10-08:
	// Usually would be Friday 2026-10-09, but since Friday is a holiday,
	// it must skip to next Wednesday: 2026-10-14!
	thursday := time.Date(2026, 10, 8, 14, 0, 0, 0, loc)
	nextHearing := cal.NextHearingDay(thursday)

	if nextHearing.Weekday() != time.Wednesday {
		t.Errorf("Next hearing should skip holiday Friday to Wednesday, got %v", nextHearing.Weekday())
	}
	if nextHearing.Day() != 14 || nextHearing.Month() != time.October {
		t.Errorf("Expected 2026-10-14, got %v", nextHearing.Format("2006-01-02"))
	}
}

// TestDeadlineCrossingMultipleHolidaysAndWeekend tests AddWorkingHours skipping weekends and consecutive holidays.
func TestDeadlineCrossingMultipleHolidaysAndWeekend(t *testing.T) {
	loc := AddisAbabaLocation()

	// Thursday 2026-10-08 at 15:00.
	// Office hours: 08:30 to 17:00 (8.5h per day).
	// Thursday remaining hours: 15:00 to 17:00 = 2 hours.
	// Let Friday 2026-10-09 and Monday 2026-10-12 be holidays.
	// Saturday 2026-10-10 & Sunday 2026-10-11 are weekend off.
	// Add 5 working hours:
	// - Thursday consumes 2 hours (15:00 to 17:00). 3 hours remain.
	// - Friday: holiday (skipped).
	// - Saturday/Sunday: weekend (skipped).
	// - Monday: holiday (skipped).
	// - Tuesday 2026-10-13: office opens at 08:30.
	//   Remaining 3 hours added to 08:30 -> Tuesday 11:30!

	holidays := map[string]bool{
		"2026-10-09": true, // Friday
		"2026-10-12": true, // Monday
	}

	checker := func(date time.Time, structureID string) bool {
		return holidays[date.Format("2006-01-02")]
	}

	cal := NewCalendar(DefaultConfig(), checker)

	start := time.Date(2026, 10, 8, 15, 0, 0, 0, loc)
	deadline := cal.AddWorkingHours(start, 5.0)

	expected := time.Date(2026, 10, 13, 11, 30, 0, 0, loc)
	if !deadline.Equal(expected) {
		t.Errorf("Deadline crossing holidays mismatch: got %v, want %v", deadline, expected)
	}

	// Working days count between start and deadline should be 5h / 8.5h
	expectedDays := 5.0 / 8.5
	cnt := cal.CountWorkingDays(start, deadline)
	diff := cnt - expectedDays
	if diff < -0.01 || diff > 0.01 {
		t.Errorf("CountWorkingDays mismatch: got %f, want %f", cnt, expectedDays)
	}
}

// TestGeezNumerals tests formatting numbers into Ge'ez numerals.
func TestGeezNumerals(t *testing.T) {
	testCases := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{-5, "-5"},
		{10000, "10000"},
		{1, "፩"},
		{7, "፯"},
		{10, "፲"},
		{17, "፲፯"},
		{23, "፳፫"},
		{100, "፻"},
		{200, "፪፻"},
		{2017, "፳፻፲፯"},
	}

	for _, tc := range testCases {
		res := ToGeezNumeral(tc.input)
		if res != tc.expected {
			t.Errorf("ToGeezNumeral(%d): got %s, want %s", tc.input, res, tc.expected)
		}
	}
}

// TestAddWorkingDays tests adding positive, zero, and negative working days.
func TestAddWorkingDays(t *testing.T) {
	loc := AddisAbabaLocation()
	cal := NewCalendar(DefaultConfig(), nil)

	// Friday 2026-10-02
	friday := time.Date(2026, 10, 2, 10, 0, 0, 0, loc)

	// 0 days
	if !cal.AddWorkingDays(friday, 0).Equal(friday) {
		t.Errorf("AddWorkingDays with 0 should return same time")
	}

	// +1 working day skips Sat & Sun -> Monday 2026-10-05
	monday := cal.AddWorkingDays(friday, 1)
	if monday.Weekday() != time.Monday || monday.Day() != 5 {
		t.Errorf("AddWorkingDays(+1) from Friday should be Monday Oct 5, got %v", monday)
	}

	// -1 working day from Monday 2026-10-05 -> Friday 2026-10-02
	prevFriday := cal.AddWorkingDays(monday, -1)
	if prevFriday.Weekday() != time.Friday || prevFriday.Day() != 2 {
		t.Errorf("AddWorkingDays(-1) from Monday should be Friday Oct 2, got %v", prevFriday)
	}
}

// TestSaturdayHalfDayConfiguration tests working calendar with Saturday enabled as half day.
func TestSaturdayHalfDayConfiguration(t *testing.T) {
	loc := AddisAbabaLocation()
	cfg := DefaultConfig()
	cfg.SaturdayHalfDay = true
	cfg.SatEndHour = 12
	cfg.SatEndMinute = 30

	cal := NewCalendar(cfg, nil)

	satHours := cal.SaturdayHours()
	if satHours != 4.0 {
		t.Errorf("Expected Saturday hours 4.0, got %f", satHours)
	}

	// Saturday 2026-10-03
	saturday := time.Date(2026, 10, 3, 9, 0, 0, 0, loc)
	if !cal.IsWorkingDay(saturday, "") {
		t.Errorf("Saturday should be working day when SaturdayHalfDay is true")
	}

	// Sunday 2026-10-04 is always off
	sunday := time.Date(2026, 10, 4, 10, 0, 0, 0, loc)
	if cal.IsWorkingDay(sunday, "") {
		t.Errorf("Sunday must never be a working day")
	}
}

// TestNextHearingSlots tests fetching multiple consecutive hearing slots.
func TestNextHearingSlots(t *testing.T) {
	loc := AddisAbabaLocation()
	cal := NewCalendar(DefaultConfig(), nil)

	// Thursday 2026-10-01
	thursday := time.Date(2026, 10, 1, 9, 0, 0, 0, loc)
	slots := cal.NextHearingSlots(thursday, 4)

	if len(slots) != 4 {
		t.Fatalf("Expected 4 slots, got %d", len(slots))
	}

	// Expected: Friday Oct 2, Wednesday Oct 7, Friday Oct 9, Wednesday Oct 14
	expectedWeekdays := []time.Weekday{time.Friday, time.Wednesday, time.Friday, time.Wednesday}
	for i, s := range slots {
		if s.Weekday() != expectedWeekdays[i] {
			t.Errorf("Slot %d: expected %v, got %v", i, expectedWeekdays[i], s.Weekday())
		}
	}

	// Default count fallback when <= 0
	defSlots := cal.NextHearingSlots(thursday, 0)
	if len(defSlots) != 5 {
		t.Errorf("Expected default 5 slots when count <= 0, got %d", len(defSlots))
	}
}

// TestEthiopianMonthNamesAndFormatting tests Amharic & English month names and string formatting.
func TestEthiopianMonthNamesAndFormatting(t *testing.T) {
	if EthiopianMonthName(0, "en") != "" || EthiopianMonthName(14, "en") != "" {
		t.Errorf("Invalid month should return empty string")
	}

	if EthiopianMonthName(1, "en") != "Meskerem" || EthiopianMonthName(1, "am") != "መስከረም" {
		t.Errorf("Month 1 name mismatch")
	}
	if EthiopianMonthName(13, "en") != "Pagume" || EthiopianMonthName(13, "am") != "ጳጉሜ" {
		t.Errorf("Month 13 name mismatch")
	}

	ed := EthiopianDate{Year: 2017, Month: 1, Day: 17}
	enStr := FormatEthiopian(ed, "en")
	amStr := FormatEthiopian(ed, "am")

	if enStr != "Meskerem 17, 2017" {
		t.Errorf("FormatEthiopian EN mismatch: got %s", enStr)
	}
	if amStr != "መስከረም 17, 2017" {
		t.Errorf("FormatEthiopian AM mismatch: got %s", amStr)
	}

	// Invalid days in month
	if DaysInEthiopianMonth(2017, 0) != 0 || DaysInEthiopianMonth(2017, 14) != 0 {
		t.Errorf("DaysInEthiopianMonth should be 0 for invalid month")
	}
}

// TestPackageLevelConvenienceFunctions tests the package-level wrappers delegating to Default().
func TestPackageLevelConvenienceFunctions(t *testing.T) {
	loc := AddisAbabaLocation()
	testChecker := func(date time.Time, structureID string) bool {
		return false
	}
	SetDefaultHolidayChecker(testChecker)

	now := time.Date(2026, 10, 2, 10, 0, 0, 0, loc)
	_ = IsWorkingDay(now, "")
	_ = AddWorkingDays(now, 1)
	_ = AddWorkingHours(now, 2.5)
	_ = CountWorkingDays(now, now.AddDate(0, 0, 3))
	_ = NextHearingDay(now)
	_ = NextHearingSlots(now, 3)

	// Inverted CountWorkingDays (from > to)
	negDays := CountWorkingDays(now.AddDate(0, 0, 3), now)
	if negDays >= 0 {
		t.Errorf("Expected negative working days when from > to, got %f", negDays)
	}

	// AddWorkingHours with <= 0
	if !AddWorkingHours(now, 0).Equal(now) {
		t.Errorf("AddWorkingHours(0) should return unchanged time")
	}

	// Calendar with invalid timezone fallback
	cfgInvalid := DefaultConfig()
	cfgInvalid.TimeZone = "NonExistent/Timezone"
	calFallback := NewCalendar(cfgInvalid, nil)
	if calFallback == nil {
		t.Fatalf("Expected valid calendar on fallback")
	}

	// Saturday hours when Saturday is off
	calSatOff := NewCalendar(DefaultConfig(), nil)
	if calSatOff.SaturdayHours() != 0 {
		t.Errorf("Saturday hours should be 0 when SaturdayHalfDay is false")
	}

	// Calendar location nil fallback
	calNilLoc := &Calendar{config: DefaultConfig()}
	if calNilLoc.Location() == nil {
		t.Errorf("Location() should fallback to AddisAbabaLocation when loc is nil")
	}

	// Office hours for Sunday (not a working day)
	sunday := time.Date(2026, 10, 4, 10, 0, 0, 0, loc)
	_, _, ok := calSatOff.OfficeHoursForDay(sunday, "")
	if ok {
		t.Errorf("Sunday should not be a working day for OfficeHoursForDay")
	}

	// Office hours for Monday (normal working day)
	monday := time.Date(2026, 10, 5, 10, 0, 0, 0, loc)
	startWork, endWork, okWork := calSatOff.OfficeHoursForDay(monday, "")
	if !okWork || startWork.Hour() != 8 || endWork.Hour() != 17 {
		t.Errorf("Monday office hours should be 08:30 - 17:00")
	}

	// Standard work day hours
	if calSatOff.StandardWorkDayHours() != 8.5 {
		t.Errorf("Expected standard 8.5 hours per working day")
	}
}

