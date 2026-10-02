package workcalendar

import (
	"sync"
	"time"
	_ "time/tzdata" // Embed tzdata so Africa/Addis_Ababa works everywhere
)

// Timezone constant
const DefaultTimeZone = "Africa/Addis_Ababa"

var (
	addisAbabaLoc     *time.Location
	addisAbabaLocOnce sync.Once
)

// AddisAbabaLocation returns the *time.Location for Africa/Addis_Ababa (UTC+3).
func AddisAbabaLocation() *time.Location {
	addisAbabaLocOnce.Do(func() {
		loc, err := time.LoadLocation(DefaultTimeZone)
		if err != nil {
			// Fallback to UTC+3 fixed zone if tzdata load failed
			loc = time.FixedZone("EAT", 3*3600)
		}
		addisAbabaLoc = loc
	})
	return addisAbabaLoc
}

// Config defines the working calendar configuration
type Config struct {
	TimeZone         string        // e.g. "Africa/Addis_Ababa"
	WorkDays         []time.Weekday // Default: Monday through Friday
	SaturdayHalfDay  bool          // If true, Saturday counts as a half working day
	StartHour        int           // Default: 8
	StartMinute      int           // Default: 30 (08:30)
	EndHour          int           // Default: 17
	EndMinute        int           // Default: 0  (17:00)
	SatEndHour       int           // Default: 12
	SatEndMinute     int           // Default: 30 (12:30 for half day)
}

// DefaultConfig returns the standard Ethiopian public service configuration:
// Monday - Friday: 08:30 - 17:00
// Saturday: off (can be toggled to half day 08:30 - 12:30)
// Sunday: always off
func DefaultConfig() Config {
	return Config{
		TimeZone:        DefaultTimeZone,
		WorkDays:        []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		SaturdayHalfDay: false,
		StartHour:       8,
		StartMinute:     30,
		EndHour:         17,
		EndMinute:       0,
		SatEndHour:      12,
		SatEndMinute:    30,
	}
}

// HolidayChecker is a function or interface that determines if a given date is a holiday.
// date is normalized to Africa/Addis_Ababa at 00:00:00.
type HolidayChecker func(date time.Time, structureID string) bool

// Calendar is the pure domain working calendar engine.
type Calendar struct {
	config         Config
	loc            *time.Location
	holidayChecker HolidayChecker
}

var (
	defaultCalendar     *Calendar
	defaultCalendarOnce sync.Once
)

// NewCalendar creates a new working calendar instance.
func NewCalendar(cfg Config, checker HolidayChecker) *Calendar {
	loc, err := time.LoadLocation(cfg.TimeZone)
	if err != nil {
		loc = AddisAbabaLocation()
	}
	if len(cfg.WorkDays) == 0 {
		cfg.WorkDays = []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}
	}
	return &Calendar{
		config:         cfg,
		loc:            loc,
		holidayChecker: checker,
	}
}

// Default returns the default global Calendar instance.
func Default() *Calendar {
	defaultCalendarOnce.Do(func() {
		defaultCalendar = NewCalendar(DefaultConfig(), nil)
	})
	return defaultCalendar
}

// SetDefaultHolidayChecker sets the global default holiday checker.
func SetDefaultHolidayChecker(checker HolidayChecker) {
	if defaultCalendar == nil {
		Default()
	}
	defaultCalendar.holidayChecker = checker
}

// Location returns the calendar's timezone location.
func (c *Calendar) Location() *time.Location {
	if c.loc != nil {
		return c.loc
	}
	return AddisAbabaLocation()
}

// NormalizeDate returns the date truncated to midnight (00:00:00) in the calendar timezone.
func (c *Calendar) NormalizeDate(t time.Time) time.Time {
	inLoc := t.In(c.Location())
	return time.Date(inLoc.Year(), inLoc.Month(), inLoc.Day(), 0, 0, 0, 0, c.Location())
}

// IsWorkingDay checks if a given date is a working day, optionally scoped to a structureID.
func (c *Calendar) IsWorkingDay(date time.Time, structureID string) bool {
	norm := c.NormalizeDate(date)
	wd := norm.Weekday()

	// Sunday is always off
	if wd == time.Sunday {
		return false
	}

	// Saturday check
	if wd == time.Saturday {
		if !c.config.SaturdayHalfDay {
			return false
		}
		// Saturday is a half day, still considered a working day unless it's a holiday
	} else {
		// Monday to Friday check
		isConfiguredWorkDay := false
		for _, day := range c.config.WorkDays {
			if wd == day {
				isConfiguredWorkDay = true
				break
			}
		}
		if !isConfiguredWorkDay {
			return false
		}
	}

	// Check holiday
	if c.holidayChecker != nil && c.holidayChecker(norm, structureID) {
		return false
	}

	return true
}

// OfficeHoursForDay returns the start and end of office hours for a given day in the calendar timezone.
// If the day is not a working day, it returns zero times and false.
func (c *Calendar) OfficeHoursForDay(date time.Time, structureID string) (start time.Time, end time.Time, ok bool) {
	if !c.IsWorkingDay(date, structureID) {
		return time.Time{}, time.Time{}, false
	}

	norm := c.NormalizeDate(date)
	wd := norm.Weekday()

	start = time.Date(norm.Year(), norm.Month(), norm.Day(), c.config.StartHour, c.config.StartMinute, 0, 0, c.Location())
	if wd == time.Saturday && c.config.SaturdayHalfDay {
		end = time.Date(norm.Year(), norm.Month(), norm.Day(), c.config.SatEndHour, c.config.SatEndMinute, 0, 0, c.Location())
	} else {
		end = time.Date(norm.Year(), norm.Month(), norm.Day(), c.config.EndHour, c.config.EndMinute, 0, 0, c.Location())
	}
	return start, end, true
}

// StandardWorkDayHours returns the normal daily office hours (e.g. 17:00 - 08:30 = 8.5 hours).
func (c *Calendar) StandardWorkDayHours() float64 {
	startMin := c.config.StartHour*60 + c.config.StartMinute
	endMin := c.config.EndHour*60 + c.config.EndMinute
	diffMin := endMin - startMin
	if diffMin <= 0 {
		return 8.0
	}
	return float64(diffMin) / 60.0
}

// SaturdayHours returns Saturday office hours (e.g. 12:30 - 08:30 = 4.0 hours).
func (c *Calendar) SaturdayHours() float64 {
	if !c.config.SaturdayHalfDay {
		return 0.0
	}
	startMin := c.config.StartHour*60 + c.config.StartMinute
	endMin := c.config.SatEndHour*60 + c.config.SatEndMinute
	diffMin := endMin - startMin
	if diffMin <= 0 {
		return 4.0
	}
	return float64(diffMin) / 60.0
}

// AddWorkingDays adds n working days to the start time.
// If n is negative, it subtracts working days.
func (c *Calendar) AddWorkingDays(start time.Time, n int) time.Time {
	if n == 0 {
		return start
	}

	curr := start.In(c.Location())
	step := 1
	remaining := n
	if n < 0 {
		step = -1
		remaining = -n
	}

	for remaining > 0 {
		curr = curr.AddDate(0, 0, step)
		if c.IsWorkingDay(curr, "") {
			remaining--
		}
	}
	return curr
}

// AddWorkingHours adds h working hours to start time, skipping weekends, holidays and non-office hours.
// If start is outside office hours (e.g. at night or weekend/holiday), it rolls forward to
// 08:30 on the next valid working day before counting.
func (c *Calendar) AddWorkingHours(start time.Time, h float64) time.Time {
	if h <= 0 {
		return start
	}

	curr := start.In(c.Location())
	remSeconds := h * 3600.0

	for remSeconds > 0 {
		dayStart, dayEnd, isWork := c.OfficeHoursForDay(curr, "")
		if !isWork {
			// Advance to next day at 00:00 and continue
			norm := c.NormalizeDate(curr)
			curr = norm.AddDate(0, 0, 1)
			continue
		}

		// If current time is before office opens, advance to dayStart
		if curr.Before(dayStart) {
			curr = dayStart
		}

		// If current time is at or after office closes, advance to next day
		if !curr.Before(dayEnd) {
			norm := c.NormalizeDate(curr)
			curr = norm.AddDate(0, 0, 1)
			continue
		}

		// Calculate available working seconds in the current business day
		availableSec := dayEnd.Sub(curr).Seconds()
		if remSeconds <= availableSec {
			return curr.Add(time.Duration(remSeconds * float64(time.Second)))
		}

		// Consume the rest of today's office hours and move to next day
		remSeconds -= availableSec
		norm := c.NormalizeDate(curr)
		curr = norm.AddDate(0, 0, 1)
	}

	return curr
}

// CountWorkingDays counts the number of working days between from and to.
// Returns fractional days based on working hours.
// If from is after to, returns negative.
func (c *Calendar) CountWorkingDays(from, to time.Time) float64 {
	fromIn := from.In(c.Location())
	toIn := to.In(c.Location())

	if fromIn.Equal(toIn) {
		return 0
	}

	sign := 1.0
	if fromIn.After(toIn) {
		fromIn, toIn = toIn, fromIn
		sign = -1.0
	}

	stdHours := c.StandardWorkDayHours()
	if stdHours <= 0 {
		stdHours = 8.5
	}

	totalWorkingSeconds := 0.0
	curr := fromIn

	for {
		norm := c.NormalizeDate(curr)
		toNorm := c.NormalizeDate(toIn)

		dayStart, dayEnd, isWork := c.OfficeHoursForDay(curr, "")
		if isWork {
			periodStart := dayStart
			if curr.After(periodStart) {
				periodStart = curr
			}

			periodEnd := dayEnd
			if norm.Equal(toNorm) && toIn.Before(periodEnd) {
				periodEnd = toIn
			}

			if periodEnd.After(periodStart) {
				totalWorkingSeconds += periodEnd.Sub(periodStart).Seconds()
			}
		}

		if norm.Equal(toNorm) {
			break
		}

		// Advance to next day at 00:00
		curr = norm.AddDate(0, 0, 1)
	}

	workingDays := (totalWorkingSeconds / 3600.0) / stdHours
	return sign * workingDays
}

// NextHearingDay returns the next Wednesday or Friday that is strictly after date and is a working day (not a holiday).
func (c *Calendar) NextHearingDay(date time.Time) time.Time {
	curr := c.NormalizeDate(date)
	for i := 0; i < 60; i++ { // search up to 60 days
		curr = curr.AddDate(0, 0, 1)
		wd := curr.Weekday()
		if wd == time.Wednesday || wd == time.Friday {
			if c.IsWorkingDay(curr, "") {
				return curr
			}
		}
	}
	return curr
}

// NextHearingSlots returns the next `count` valid hearing dates (Wednesdays and Fridays that are working days),
// starting from `from` (inclusive if `from` is a hearing day, or strictly after).
func (c *Calendar) NextHearingSlots(from time.Time, count int) []time.Time {
	if count <= 0 {
		count = 5
	}
	slots := make([]time.Time, 0, count)
	curr := c.NormalizeDate(from)

	// If `from` is already a valid hearing day, include it
	wd := curr.Weekday()
	if (wd == time.Wednesday || wd == time.Friday) && c.IsWorkingDay(curr, "") {
		slots = append(slots, curr)
	}

	for len(slots) < count {
		curr = curr.AddDate(0, 0, 1)
		currWd := curr.Weekday()
		if (currWd == time.Wednesday || currWd == time.Friday) && c.IsWorkingDay(curr, "") {
			slots = append(slots, curr)
		}
	}
	return slots
}

// Package-level convenience functions delegating to Default():

func IsWorkingDay(date time.Time, structureID string) bool {
	return Default().IsWorkingDay(date, structureID)
}

func AddWorkingDays(start time.Time, n int) time.Time {
	return Default().AddWorkingDays(start, n)
}

func AddWorkingHours(start time.Time, h float64) time.Time {
	return Default().AddWorkingHours(start, h)
}

func CountWorkingDays(from, to time.Time) float64 {
	return Default().CountWorkingDays(from, to)
}

func NextHearingDay(date time.Time) time.Time {
	return Default().NextHearingDay(date)
}

func NextHearingSlots(from time.Time, count int) []time.Time {
	return Default().NextHearingSlots(from, count)
}
