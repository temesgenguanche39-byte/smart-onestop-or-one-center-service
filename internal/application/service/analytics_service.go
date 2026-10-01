package service

import (
	"context"
	"fmt"
	"time"

	"github.com/smart-onestop/platform/internal/application/dto"
	"github.com/smart-onestop/platform/internal/domain"
)

type AnalyticsService struct {
	caseRepo      domain.CaseRepository
	structureRepo domain.StructureRepository
	auditRepo     domain.AuditRepository
}

func NewAnalyticsService(
	caseRepo domain.CaseRepository,
	structureRepo domain.StructureRepository,
	auditRepo domain.AuditRepository,
) *AnalyticsService {
	return &AnalyticsService{
		caseRepo:      caseRepo,
		structureRepo: structureRepo,
		auditRepo:     auditRepo,
	}
}

func (s *AnalyticsService) GetExecutiveSummary(ctx context.Context, structureID *uint) (*dto.AnalyticsSummaryResponse, error) {
	allCases, _, err := s.caseRepo.List(ctx, domain.CaseFilter{
		StructureID: structureID,
		Limit:       2000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cases for analytics: %w", err)
	}

	now := time.Now().UTC()
	var (
		totalCases        int64
		activeCases       int64
		resolvedCases     int64
		underReviewCases  int64
		scheduledHearings int64
		breachedCases     int64
		escalatedSubCity  int64
		escalatedCity     int64
		totalResolveHrs   float64
		resolvedWithTime  int64
	)

	totalCases = int64(len(allCases))

	// Structure metrics map
	structMap := make(map[uint]*dto.StructureMetric)

	for _, c := range allCases {
		// Status counters
		switch c.Status {
		case domain.StatusResolved:
			resolvedCases++
			if c.ResolvedAt != nil {
				hrs := c.ResolvedAt.Sub(c.CreatedAt).Hours()
				if hrs >= 0 {
					totalResolveHrs += hrs
					resolvedWithTime++
				}
			}
		case domain.StatusUnderReview:
			underReviewCases++
			activeCases++
		case domain.StatusScheduledForHearing:
			scheduledHearings++
			activeCases++
		case domain.StatusEscalatedToSubCity:
			escalatedSubCity++
			activeCases++
		case domain.StatusEscalatedToCity:
			escalatedCity++
			activeCases++
		default:
			activeCases++
		}

		isBreached := c.SLADeadline.Before(now) && c.Status != domain.StatusResolved && c.Status != domain.StatusRejected
		if isBreached {
			breachedCases++
		}

		// Group by structure
		if _, exists := structMap[c.CurrentStructureID]; !exists {
			name := "Unknown Tier"
			level := domain.AdminLevelWoreda
			if c.CurrentStructure != nil {
				name = c.CurrentStructure.Name
				level = c.CurrentStructure.Level
			}
			structMap[c.CurrentStructureID] = &dto.StructureMetric{
				StructureID: c.CurrentStructureID,
				Name:        name,
				Level:       level,
			}
		}
		sm := structMap[c.CurrentStructureID]
		sm.CaseCount++
		if isBreached {
			sm.BreachedCount++
		}
		if c.Status == domain.StatusResolved {
			sm.ResolvedCount++
		}
	}

	avgHours := 0.0
	if resolvedWithTime > 0 {
		avgHours = totalResolveHrs / float64(resolvedWithTime)
	}

	var breakdown []dto.StructureMetric
	for _, sm := range structMap {
		breakdown = append(breakdown, *sm)
	}

	// Fetch recent audit logs
	recentLogs, _ := s.auditRepo.ListRecent(ctx, 15)
	var logDTOs []dto.AuditLogResponse
	for _, l := range recentLogs {
		perfName := "System Engine"
		if l.Performer != nil {
			perfName = l.Performer.FullName
		}
		logDTOs = append(logDTOs, dto.AuditLogResponse{
			ID:            l.ID,
			CaseID:        l.CaseID,
			Action:        l.Action,
			PerformerName: perfName,
			OldValue:      l.OldValue,
			NewValue:      l.NewValue,
			Notes:         l.Notes,
			CreatedAt:     l.CreatedAt,
		})
	}

	return &dto.AnalyticsSummaryResponse{
		TotalCases:           totalCases,
		ActiveCases:          activeCases,
		ResolvedCases:        resolvedCases,
		UnderReviewCases:     underReviewCases,
		ScheduledHearings:    scheduledHearings,
		BreachedCases:        breachedCases,
		EscalatedToSubCity:   escalatedSubCity,
		EscalatedToCity:      escalatedCity,
		AverageResolutionHrs: avgHours,
		StructureBreakdown:   breakdown,
		RecentAuditLogs:      logDTOs,
	}, nil
}
