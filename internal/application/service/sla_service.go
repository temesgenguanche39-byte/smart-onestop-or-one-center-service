package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/smart-onestop/platform/internal/domain"
)

type SLAService struct {
	caseRepo      domain.CaseRepository
	structureRepo domain.StructureRepository
	auditRepo     domain.AuditRepository
}

func NewSLAService(
	caseRepo domain.CaseRepository,
	structureRepo domain.StructureRepository,
	auditRepo domain.AuditRepository,
) *SLAService {
	return &SLAService{
		caseRepo:      caseRepo,
		structureRepo: structureRepo,
		auditRepo:     auditRepo,
	}
}

// SweepAndEscalate executes the race-condition-free SLA breach auto-escalation
func (s *SLAService) SweepAndEscalate(ctx context.Context) (int, error) {
	now := time.Now().UTC()

	// 1. Fetch breached cases with row-level locking (SKIP LOCKED)
	breachedCases, err := s.caseRepo.FetchAndLockBreachedCases(ctx, now, 100)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch breached cases: %w", err)
	}

	if len(breachedCases) == 0 {
		return 0, nil
	}

	escalatedCount := 0

	for i := range breachedCases {
		c := &breachedCases[i]

		// Resolve current administrative structure
		currentStruct, err := s.structureRepo.GetByID(ctx, c.CurrentStructureID)
		if err != nil || currentStruct.ParentID == nil {
			// Already at top municipal tier (City) or unlinked; log warning
			_ = s.auditRepo.Log(ctx, &domain.CaseAuditLog{
				CaseID:    c.ID,
				Action:    "SLA_BREACH_TOP_TIER_ALERT",
				Notes:     "Case breached SLA at City Administration level. Executive intervention required.",
				CreatedAt: now,
			})
			continue
		}

		parentStruct, err := s.structureRepo.GetByID(ctx, *currentStruct.ParentID)
		if err != nil {
			log.Printf("[SLA Worker] Parent structure %d not found for case %s", *currentStruct.ParentID, c.TicketNumber)
			continue
		}

		oldStatus := c.Status
		c.CurrentStructureID = parentStruct.ID
		c.IsEscalated = true
		c.EscalationCount++
		c.SLADeadline = now.Add(24 * time.Hour) // Mandatory +24h extension per municipal policy
		c.UpdatedAt = now

		actionName := "AUTO_ESCALATED_SLA_BREACH"
		notes := ""

		if currentStruct.Level == domain.AdminLevelWoreda {
			c.Status = domain.StatusEscalatedToSubCity
			actionName = "AUTO_ESCALATED_TO_SUBCITY"
			notes = fmt.Sprintf("Automated SLA breach transition: %s (Woreda) -> %s (Sub-City)", currentStruct.Name, parentStruct.Name)
		} else if currentStruct.Level == domain.AdminLevelSubCity {
			c.Status = domain.StatusEscalatedToCity
			actionName = "AUTO_ESCALATED_TO_CITY"
			notes = fmt.Sprintf("Automated SLA breach transition: %s (Sub-City) -> %s (City Administration)", currentStruct.Name, parentStruct.Name)
		}

		if err := s.caseRepo.Update(ctx, c); err != nil {
			log.Printf("[SLA Worker] Failed to update case %s during auto-escalation: %v", c.TicketNumber, err)
			continue
		}

		oldVal, _ := json.Marshal(map[string]interface{}{
			"status":    oldStatus,
			"structure": currentStruct.Name,
			"level":     currentStruct.Level,
		})
		newVal, _ := json.Marshal(map[string]interface{}{
			"status":       c.Status,
			"structure":    parentStruct.Name,
			"level":        parentStruct.Level,
			"new_deadline": c.SLADeadline,
		})

		// Append-only audit record (System Actor: PerformedBy = nil)
		_ = s.auditRepo.Log(ctx, &domain.CaseAuditLog{
			CaseID:      c.ID,
			PerformedBy: nil, // System automation actor
			Action:      actionName,
			OldValue:    string(oldVal),
			NewValue:    string(newVal),
			Notes:       notes,
			CreatedAt:   now,
		})

		escalatedCount++
		log.Printf("[SLA Engine] Case %s successfully auto-escalated to %s (%s)", c.TicketNumber, parentStruct.Name, c.Status)
	}

	return escalatedCount, nil
}
