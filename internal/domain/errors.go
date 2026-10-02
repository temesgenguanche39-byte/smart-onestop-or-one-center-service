package domain

import "errors"

var (
	ErrCaseNotFound           = errors.New("case ticket not found")
	ErrCitizenNotFound        = errors.New("citizen not found")
	ErrStructureNotFound      = errors.New("administrative structure not found")
	ErrUserNotFound           = errors.New("user not found")
	ErrUnauthorized           = errors.New("unauthorized operation")
	ErrForbidden              = errors.New("forbidden: insufficient administrative permissions")
	ErrInvalidCredentials     = errors.New("invalid email or password")
	ErrCaseAlreadyResolved    = errors.New("case has already been resolved or rejected")
	ErrSLANotBreached         = errors.New("case SLA deadline has not been breached")
	ErrAlreadyTopTier         = errors.New("case is already at the highest municipal tier (City Administration)")
	ErrNotWednesdayOrFriday   = errors.New("hearing slots can only be scheduled on Wednesday or Friday")
	ErrHearingOnHoliday       = errors.New("hearing cannot be scheduled on a public holiday")
	ErrHolidayNotFound        = errors.New("holiday not found")
	ErrInvalidDateFormat      = errors.New("invalid date format")
	ErrSlotAlreadyBooked      = errors.New("the selected hearing time slot is already booked")
	ErrInvalidSlotTime        = errors.New("invalid hearing slot duration or time format")
	ErrDuplicateTicket        = errors.New("case with this ticket number already exists")
	ErrInvalidHierarchyLevel  = errors.New("invalid administrative hierarchy tier")
)
