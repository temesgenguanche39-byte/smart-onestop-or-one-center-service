package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/smart-onestop/platform/internal/application/dto"
	"github.com/smart-onestop/platform/internal/domain"
)

type CaseService struct {
	caseRepo        domain.CaseRepository
	citizenRepo     domain.CitizenRepository
	structureRepo   domain.StructureRepository
	serviceTypeRepo domain.ServiceTypeRepository
	auditRepo       domain.AuditRepository
	hearingRepo     domain.HearingRepository
}

func NewCaseService(
	caseRepo domain.CaseRepository,
	citizenRepo domain.CitizenRepository,
	structureRepo domain.StructureRepository,
	serviceTypeRepo domain.ServiceTypeRepository,
	auditRepo domain.AuditRepository,
	hearingRepo domain.HearingRepository,
) *CaseService {
	return &CaseService{
		caseRepo:        caseRepo,
		citizenRepo:     citizenRepo,
		structureRepo:   structureRepo,
		serviceTypeRepo: serviceTypeRepo,
		auditRepo:       auditRepo,
		hearingRepo:     hearingRepo,
	}
}

// CreateCase handles citizen grievance intake
func (s *CaseService) CreateCase(ctx context.Context, req dto.CreateCaseRequest) (*dto.CaseResponse, error) {
	// 1. Verify structure (Woreda)
	structNode, err := s.structureRepo.GetByID(ctx, req.WoredaID)
	if err != nil {
		return nil, domain.ErrStructureNotFound
	}

	// 2. Fetch service type for base SLA hours
	serviceType, err := s.serviceTypeRepo.GetByID(ctx, req.ServiceTypeID)
	if err != nil {
		return nil, fmt.Errorf("service type not found: %w", err)
	}

	// 3. Resolve or register citizen
	citizen, err := s.citizenRepo.GetByPhone(ctx, req.CitizenPhone)
	if err != nil {
		citizen = &domain.Citizen{
			ID:                uuid.New(),
			FullName:          req.CitizenFullName,
			PhoneNumber:       req.CitizenPhone,
			NationalID:        req.CitizenNationalID,
			WoredaID:          &req.WoredaID,
			HouseNumber:       req.CitizenHouseNumber,
			PreferredLanguage: "am",
			CreatedAt:         time.Now().UTC(),
		}
		if err := s.citizenRepo.Create(ctx, citizen); err != nil {
			return nil, fmt.Errorf("failed to register citizen: %w", err)
		}
	}

	// 4. Generate Ticket Number & QR verification code
	ticketNum, err := generateTicketNumber()
	if err != nil {
		return nil, err
	}
	qrPayload := generateQRCodeHash(ticketNum)

	now := time.Now().UTC()
	slaHours := serviceType.BaseSLAHours
	if slaHours <= 0 {
		slaHours = 48
	}
	slaDeadline := now.Add(time.Duration(slaHours) * time.Hour)

	priority := req.Priority
	if priority == "" {
		priority = domain.PriorityNormal
	}

	caseEntity := &domain.Case{
		ID:                 uuid.New(),
		TicketNumber:       ticketNum,
		CitizenID:          citizen.ID,
		Citizen:            citizen,
		ServiceTypeID:      serviceType.ID,
		ServiceType:        serviceType,
		CurrentStructureID: structNode.ID,
		CurrentStructure:   structNode,
		Title:              req.Title,
		Description:        req.Description,
		Status:             domain.StatusSubmitted,
		Priority:           priority,
		SLADeadline:        slaDeadline,
		IsEscalated:        false,
		EscalationCount:    0,
		QRVerificationCode: qrPayload,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	// Process attachments if any
	for _, att := range req.Attachments {
		caseEntity.Attachments = append(caseEntity.Attachments, domain.CaseAttachment{
			ID:               uuid.New(),
			CaseID:           caseEntity.ID,
			FileName:         att.FileName,
			FileURL:          att.FileURL,
			MIMEType:         att.MIMEType,
			ExtractedOCRText: att.ExtractedOCRText,
			UploadedAt:       now,
		})
	}

	if err := s.caseRepo.Create(ctx, caseEntity); err != nil {
		return nil, fmt.Errorf("failed to save case: %w", err)
	}

	// 5. Append-only Immutable Audit Log
	initData, _ := json.Marshal(map[string]interface{}{
		"status":        caseEntity.Status,
		"woreda":        structNode.Name,
		"sla_hours":     slaHours,
		"ticket_number": ticketNum,
	})
	audit := &domain.CaseAuditLog{
		CaseID:    caseEntity.ID,
		Action:    "CASE_INTAKE_SUBMITTED",
		NewValue:  string(initData),
		Notes:     fmt.Sprintf("Citizen %s submitted grievance via One-Stop Portal", citizen.FullName),
		CreatedAt: now,
	}
	_ = s.auditRepo.Log(ctx, audit)

	res := mapToCaseResponse(caseEntity)
	return &res, nil
}

// GetByTicket public query
func (s *CaseService) GetByTicket(ctx context.Context, ticket string) (*dto.CaseDetailResponse, error) {
	c, err := s.caseRepo.GetByTicketNumber(ctx, ticket)
	if err != nil {
		return nil, domain.ErrCaseNotFound
	}
	return s.buildCaseDetailResponse(ctx, c)
}

// GetByQR verification
func (s *CaseService) GetByQR(ctx context.Context, qrCode string) (*dto.CaseDetailResponse, error) {
	c, err := s.caseRepo.GetByQRCode(ctx, qrCode)
	if err != nil {
		return nil, domain.ErrCaseNotFound
	}
	return s.buildCaseDetailResponse(ctx, c)
}

// GetByID internal query
func (s *CaseService) GetByID(ctx context.Context, id uuid.UUID) (*dto.CaseDetailResponse, error) {
	c, err := s.caseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrCaseNotFound
	}
	return s.buildCaseDetailResponse(ctx, c)
}

// ListCases filtered by structure, status, breach
func (s *CaseService) ListCases(ctx context.Context, filter domain.CaseFilter) ([]dto.CaseResponse, int64, error) {
	cases, total, err := s.caseRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	var results []dto.CaseResponse
	for _, c := range cases {
		results = append(results, mapToCaseResponse(&c))
	}
	return results, total, nil
}

// AssignCase assigns official to ticket
func (s *CaseService) AssignCase(ctx context.Context, caseID uuid.UUID, officialID uuid.UUID, performerID uuid.UUID) error {
	c, err := s.caseRepo.GetByID(ctx, caseID)
	if err != nil {
		return domain.ErrCaseNotFound
	}

	oldUser := ""
	if c.AssignedToUserID != nil {
		oldUser = c.AssignedToUserID.String()
	}

	c.AssignedToUserID = &officialID
	if c.Status == domain.StatusSubmitted {
		c.Status = domain.StatusUnderReview
	}
	c.UpdatedAt = time.Now().UTC()

	if err := s.caseRepo.Update(ctx, c); err != nil {
		return err
	}

	oldVal, _ := json.Marshal(map[string]interface{}{"assigned_to": oldUser, "status": domain.StatusSubmitted})
	newVal, _ := json.Marshal(map[string]interface{}{"assigned_to": officialID.String(), "status": c.Status})

	_ = s.auditRepo.Log(ctx, &domain.CaseAuditLog{
		CaseID:      c.ID,
		PerformedBy: &performerID,
		Action:      "CASE_ASSIGNED_TO_OFFICER",
		OldValue:    string(oldVal),
		NewValue:    string(newVal),
		Notes:       "Case assigned for administrative review",
		CreatedAt:   time.Now().UTC(),
	})
	return nil
}

// ResolveCase closes the case with official resolution
func (s *CaseService) ResolveCase(ctx context.Context, caseID uuid.UUID, officialID uuid.UUID, req dto.ResolveCaseRequest) error {
	c, err := s.caseRepo.GetByID(ctx, caseID)
	if err != nil {
		return domain.ErrCaseNotFound
	}

	if c.Status == domain.StatusResolved || c.Status == domain.StatusRejected {
		return domain.ErrCaseAlreadyResolved
	}

	now := time.Now().UTC()
	oldStatus := c.Status
	c.Status = domain.StatusResolved
	c.ResolutionSummary = req.ResolutionSummary
	c.ResolvedBy = &officialID
	c.ResolvedAt = &now
	c.UpdatedAt = now

	if err := s.caseRepo.Update(ctx, c); err != nil {
		return err
	}

	oldVal, _ := json.Marshal(map[string]interface{}{"status": oldStatus})
	newVal, _ := json.Marshal(map[string]interface{}{
		"status":     domain.StatusResolved,
		"resolution": req.ResolutionSummary,
	})

	_ = s.auditRepo.Log(ctx, &domain.CaseAuditLog{
		CaseID:      c.ID,
		PerformedBy: &officialID,
		Action:      "CASE_RESOLVED",
		OldValue:    string(oldVal),
		NewValue:    string(newVal),
		Notes:       fmt.Sprintf("Case officially resolved. %s", req.DecisionNotes),
		CreatedAt:   now,
	})
	return nil
}

// ManualEscalate allows supervisor to manually escalate ticket
func (s *CaseService) ManualEscalate(ctx context.Context, caseID uuid.UUID, performerID uuid.UUID, reason string) error {
	c, err := s.caseRepo.GetByID(ctx, caseID)
	if err != nil {
		return domain.ErrCaseNotFound
	}

	currentStruct, err := s.structureRepo.GetByID(ctx, c.CurrentStructureID)
	if err != nil || currentStruct.ParentID == nil {
		return domain.ErrAlreadyTopTier
	}

	parentStruct, err := s.structureRepo.GetByID(ctx, *currentStruct.ParentID)
	if err != nil {
		return domain.ErrStructureNotFound
	}

	now := time.Now().UTC()
	oldStatus := c.Status
	oldStructName := currentStruct.Name

	c.CurrentStructureID = parentStruct.ID
	c.IsEscalated = true
	c.EscalationCount++
	c.SLADeadline = now.Add(24 * time.Hour) // Extend SLA by 24h
	c.UpdatedAt = now

	if currentStruct.Level == domain.AdminLevelWoreda {
		c.Status = domain.StatusEscalatedToSubCity
	} else if currentStruct.Level == domain.AdminLevelSubCity {
		c.Status = domain.StatusEscalatedToCity
	}

	if err := s.caseRepo.Update(ctx, c); err != nil {
		return err
	}

	oldVal, _ := json.Marshal(map[string]interface{}{"structure": oldStructName, "status": oldStatus})
	newVal, _ := json.Marshal(map[string]interface{}{"structure": parentStruct.Name, "status": c.Status})

	_ = s.auditRepo.Log(ctx, &domain.CaseAuditLog{
		CaseID:      c.ID,
		PerformedBy: &performerID,
		Action:      "MANUAL_ESCALATION",
		OldValue:    string(oldVal),
		NewValue:    string(newVal),
		Notes:       fmt.Sprintf("Manual escalation: %s", reason),
		CreatedAt:   now,
	})
	return nil
}

func (s *CaseService) buildCaseDetailResponse(ctx context.Context, c *domain.Case) (*dto.CaseDetailResponse, error) {
	base := mapToCaseResponse(c)
	resp := &dto.CaseDetailResponse{
		CaseResponse:      base,
		Description:       c.Description,
		ResolutionSummary: c.ResolutionSummary,
		Attachments:       c.Attachments,
	}

	if c.ResolvedByUser != nil {
		resp.ResolvedByName = c.ResolvedByUser.FullName
	}
	resp.ResolvedAt = c.ResolvedAt

	// Load audit logs
	logs, err := s.auditRepo.GetByCaseID(ctx, c.ID)
	if err == nil {
		for _, l := range logs {
			performer := "SYSTEM AUTOMATION"
			if l.Performer != nil {
				performer = fmt.Sprintf("%s (%s)", l.Performer.FullName, l.Performer.Role)
			}
			resp.AuditLogs = append(resp.AuditLogs, dto.AuditLogResponse{
				ID:            l.ID,
				CaseID:        l.CaseID,
				Action:        l.Action,
				PerformerName: performer,
				OldValue:      l.OldValue,
				NewValue:      l.NewValue,
				Notes:         l.Notes,
				CreatedAt:     l.CreatedAt,
			})
		}
	}

	// Load hearings
	hearings, err := s.hearingRepo.GetByCaseID(ctx, c.ID)
	if err == nil {
		for _, h := range hearings {
			offName := "Assigned Official"
			if h.Official != nil {
				offName = h.Official.FullName
			}
			resp.HearingSlots = append(resp.HearingSlots, dto.HearingResponse{
				ID:           h.ID,
				CaseID:       h.CaseID,
				OfficialID:   h.OfficialID,
				OfficialName: offName,
				HearingDate:  h.HearingDate.Format("2006-01-02"),
				StartTime:    h.StartTime,
				EndTime:      h.EndTime,
				HearingType:  h.HearingType,
				MeetingLink:  h.MeetingLink,
				Status:       h.Status,
				HearingNotes: h.HearingNotes,
			})
		}
	}

	return resp, nil
}

func mapToCaseResponse(c *domain.Case) dto.CaseResponse {
	now := time.Now().UTC()
	diff := c.SLADeadline.Sub(now).Hours()
	isBreached := diff < 0 && c.Status != domain.StatusResolved && c.Status != domain.StatusRejected

	citizenName := ""
	citizenPhone := ""
	if c.Citizen != nil {
		citizenName = c.Citizen.FullName
		citizenPhone = c.Citizen.PhoneNumber
	}

	serviceName := "General Administrative"
	if c.ServiceType != nil {
		serviceName = c.ServiceType.Name
	}

	structName := ""
	level := domain.AdminLevelWoreda
	if c.CurrentStructure != nil {
		structName = c.CurrentStructure.Name
		level = c.CurrentStructure.Level
	}

	assignedName := ""
	if c.AssignedToUser != nil {
		assignedName = c.AssignedToUser.FullName
	}

	return dto.CaseResponse{
		ID:                 c.ID,
		TicketNumber:       c.TicketNumber,
		CitizenName:        citizenName,
		CitizenPhone:       citizenPhone,
		ServiceTypeName:    serviceName,
		CurrentStructureID: c.CurrentStructureID,
		CurrentStructure:   structName,
		StructureLevel:     level,
		AssignedToName:     assignedName,
		Title:              c.Title,
		Status:             c.Status,
		Priority:           c.Priority,
		SLADeadline:        c.SLADeadline,
		RemainingHours:     diff,
		IsBreached:         isBreached,
		IsEscalated:        c.IsEscalated,
		EscalationCount:    c.EscalationCount,
		QRVerificationCode: c.QRVerificationCode,
		CreatedAt:          c.CreatedAt,
	}
}

func generateTicketNumber() (string, error) {
	year := time.Now().Year()
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("TKT-%d-%06d", year, n.Int64()+100000), nil
}

func generateQRCodeHash(ticket string) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("ETH-MUNI-%s-%s", ticket, hex.EncodeToString(b))
}
