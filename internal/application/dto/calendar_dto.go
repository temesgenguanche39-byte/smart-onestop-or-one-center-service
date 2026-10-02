package dto

import "github.com/smart-onestop/platform/internal/domain"

type WorkingDaysRequest struct {
	From     string `form:"from" binding:"required"` // Date string (YYYY-MM-DD)
	To       string `form:"to" binding:"required"`   // Date string (YYYY-MM-DD)
	Calendar string `form:"calendar"`                // "ethiopian" or "gregorian", default gregorian
}

type WorkingDaysResponse struct {
	FromGregorian       string  `json:"from_gregorian"`
	ToGregorian         string  `json:"to_gregorian"`
	FromEthiopian       string  `json:"from_ethiopian"`
	ToEthiopian         string  `json:"to_ethiopian"`
	FromEthiopianAM     string  `json:"from_ethiopian_am"`
	ToEthiopianAM       string  `json:"to_ethiopian_am"`
	WorkingDays         float64 `json:"working_days"`
	TotalCalendarDays   int     `json:"total_calendar_days"`
	HolidaysEncountered int     `json:"holidays_encountered"`
	WeekendDaysSkipped  int     `json:"weekend_days_skipped"`
}

type HolidayResponse struct {
	ID                 uint               `json:"id"`
	Date               string             `json:"date"` // YYYY-MM-DD
	EthiopianDate      string             `json:"ethiopian_date"`
	EthiopianDateAM    string             `json:"ethiopian_date_am"`
	NameEN             string             `json:"name_en"`
	NameAM             string             `json:"name_am"`
	Type               domain.HolidayType `json:"type"`
	Active             bool               `json:"active"`
	Source             string             `json:"source"`
	Year               int                `json:"year"`
	StructureID        *uint              `json:"structure_id,omitempty"`
}

type CreateHolidayRequest struct {
	Date        string             `json:"date" binding:"required"` // YYYY-MM-DD
	NameEN      string             `json:"name_en" binding:"required"`
	NameAM      string             `json:"name_am" binding:"required"`
	Type        domain.HolidayType `json:"type" binding:"required"`
	Active      *bool              `json:"active"`
	Source      string             `json:"source"`
	Year        int                `json:"year"`
	StructureID *uint              `json:"structure_id"`
}

type UpdateHolidayRequest struct {
	Date        string             `json:"date"`
	NameEN      string             `json:"name_en"`
	NameAM      string             `json:"name_am"`
	Type        domain.HolidayType `json:"type"`
	Active      *bool              `json:"active"`
	Source      string             `json:"source"`
	Year        int                `json:"year"`
	StructureID *uint              `json:"structure_id"`
}

type HearingSlotOption struct {
	Date            string `json:"date"`               // YYYY-MM-DD
	EthiopianDate   string `json:"ethiopian_date"`    // Meskerem 17, 2017
	EthiopianDateAM string `json:"ethiopian_date_am"` // መስከረም 17, 2017
	DayOfWeek       string `json:"day_of_week"`       // Wednesday or Friday
	DayOfWeekAM     string `json:"day_of_week_am"`    // ረቡዕ or ዓርብ
	IsWorkingDay    bool   `json:"is_working_day"`
}

type NextHearingSlotsResponse struct {
	From  string              `json:"from"`
	Count int                 `json:"count"`
	Slots []HearingSlotOption `json:"slots"`
}
