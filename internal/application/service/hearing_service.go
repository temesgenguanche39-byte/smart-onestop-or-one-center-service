package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smart-onestop/platform/internal/application/dto"
	"github.com/smart-onestop/platform/internal/domain"
)

type HearingService struct {
	hearingRepo domain.HearingRepository
	caseRepo    domain.CaseRepository
	auditRepo   domain.AuditRepository
	userRepo    domain.UserRepository
}

func NewHearingService(
	hearingRepo domain.HearingRepository,
	caseRepo domain.CaseRepository,
	auditRepo domain.AuditRepository,
	userRepo domain.UserRepository,
) *HearingService {
	return &HearingService{
		hearingRepo: hearingRepo,
		caseRepo:    caseRepo,
		auditRepo:   auditRepo,
		userRepo:    userRepo,
	}
}

// ScheduleSlot creates a hearing slot enforcing Wednesday & Friday rules
func (s *HearingService) ScheduleSlot(
	ctx context.Context,
	caseID uuid.UUID,
	officialID uuid.UUID,
	req dto.ScheduleHearingRequest,
) (*dto.HearingResponse, error) {
	// 1. Verify case
	c, err := s.caseRepo.GetByID(ctx, caseID)
	if err != nil {
		return nil, domain.ErrCaseNotFound
	}
	if c.Status == domain.StatusResolved || c.Status == domain.StatusRejected {
		return nil, domain.ErrCaseAlreadyResolved
	}

	// 2. Parse and validate hearing date
	parsedDate, err := time.Parse("2006-01-02", req.HearingDate)
	if err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	// Constraint: Must be Wednesday (3) or Friday (5)
	weekday := parsedDate.Weekday()
	if weekday != time.Wednesday && weekday != time.Friday {
		return nil, domain.ErrNotWednesdayOrFriday
	}

	// 3. Check for conflicting bookings for this official
	existingSlots, err := s.hearingRepo.ListByDate(ctx, parsedDate, &officialID)
	if err == nil {
		for _, es := range existingSlots {
			if es.StartTime == req.StartTime && es.Status == domain.HearingStatusScheduled {
				return nil, domain.ErrSlotAlreadyBooked
			}
		}
	}

	meetingLink := req.MeetingLink
	if meetingLink == "" {
		meetingLink = fmt.Sprintf("https://meet.jit.si/eth-muni-hearing-%s", c.TicketNumber)
	}

	hearingType := req.HearingType
	if hearingType == "" {
		hearingType = domain.HearingTypeVirtualCall
	}

	slot := &domain.HearingSlot{
		ID:           uuid.New(),
		CaseID:       caseID,
		OfficialID:   officialID,
		HearingDate:  parsedDate,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		HearingType:  hearingType,
		MeetingLink:  meetingLink,
		Status:       domain.HearingStatusScheduled,
		HearingNotes: req.Notes,
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.hearingRepo.CreateSlot(ctx, slot); err != nil {
		return nil, fmt.Errorf("failed to create hearing slot: %w", err)
	}

	// 4. Update case status to SCHEDULED_FOR_HEARING
	c.Status = domain.StatusScheduledForHearing
	c.UpdatedAt = time.Now().UTC()
	_ = s.caseRepo.Update(ctx, c)

	// 5. Audit Log
	newVal, _ := json.Marshal(map[string]interface{}{
		"hearing_id":   slot.ID,
		"date":         req.HearingDate,
		"time":         fmt.Sprintf("%s - %s", req.StartTime, req.EndTime),
		"meeting_link": meetingLink,
	})
	_ = s.auditRepo.Log(ctx, &domain.CaseAuditLog{
		CaseID:      c.ID,
		PerformedBy: &officialID,
		Action:      "HEARING_SCHEDULED",
		NewValue:    string(newVal),
		Notes:       fmt.Sprintf("Hearing scheduled for %s (%s)", req.HearingDate, weekday.String()),
		CreatedAt:   time.Now().UTC(),
	})

	official, _ := s.userRepo.GetByID(ctx, officialID)
	offName := "Official"
	if official != nil {
		offName = official.FullName
	}

	return &dto.HearingResponse{
		ID:           slot.ID,
		CaseID:       caseID,
		TicketNumber: c.TicketNumber,
		OfficialID:   officialID,
		OfficialName: offName,
		HearingDate:  req.HearingDate,
		StartTime:    slot.StartTime,
		EndTime:      slot.EndTime,
		HearingType:  slot.HearingType,
		MeetingLink:  slot.MeetingLink,
		Status:       slot.Status,
		HearingNotes: slot.HearingNotes,
	}, nil
}

// CompleteHearing marks slot completed and updates notes
func (s *HearingService) CompleteHearing(ctx context.Context, slotID uuid.UUID, officialID uuid.UUID, notes string) error {
	slot, err := s.hearingRepo.GetByID(ctx, slotID)
	if err != nil {
		return fmt.Errorf("hearing slot not found: %w", err)
	}

	slot.Status = domain.HearingStatusCompleted
	slot.HearingNotes = notes
	if err := s.hearingRepo.UpdateSlot(ctx, slot); err != nil {
		return err
	}

	_ = s.auditRepo.Log(ctx, &domain.CaseAuditLog{
		CaseID:      slot.CaseID,
		PerformedBy: &officialID,
		Action:      "HEARING_COMPLETED",
		Notes:       fmt.Sprintf("Digital hearing completed. Notes: %s", notes),
		CreatedAt:   time.Now().UTC(),
	})
	return nil
}
