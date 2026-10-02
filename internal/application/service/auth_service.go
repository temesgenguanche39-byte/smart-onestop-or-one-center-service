package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/smart-onestop/platform/internal/application/dto"
	"github.com/smart-onestop/platform/internal/domain"
	"github.com/smart-onestop/platform/internal/infrastructure/security"
)

type AuthService struct {
	userRepo      domain.UserRepository
	structureRepo domain.StructureRepository
	jwtManager    *security.JWTManager
}

func NewAuthService(
	userRepo domain.UserRepository,
	structureRepo domain.StructureRepository,
	jwtManager *security.JWTManager,
) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		structureRepo: structureRepo,
		jwtManager:    jwtManager,
	}
}

// Authenticate checks user credentials and issues a signed JWT
func (s *AuthService) Authenticate(ctx context.Context, email, password string) (*dto.LoginResponse, error) {
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !u.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}

	if !security.CheckPasswordHash(password, u.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	token, expiresAt, err := s.jwtManager.GenerateToken(u)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	structName, adminLevel, subCity, woreda := s.resolveJurisdictionDetails(ctx, u.StructureID)

	return &dto.LoginResponse{
		Token: token,
		User: dto.UserDTO{
			ID:            u.ID,
			FullName:      u.FullName,
			Email:         u.Email,
			PhoneNumber:   u.PhoneNumber,
			Role:          u.Role,
			StructureID:   u.StructureID,
			StructureName: structName,
			AdminLevel:    adminLevel,
			SubCityName:   subCity,
			WoredaName:    woreda,
		},
		ExpiresAt: expiresAt,
	}, nil
}

func (s *AuthService) resolveJurisdictionDetails(ctx context.Context, structureID *uint) (structName string, adminLevel domain.AdminLevel, subCityName string, woredaName string) {
	if structureID == nil {
		return "Addis Ababa City Administration (Global)", domain.AdminLevelCity, "", ""
	}
	st, err := s.structureRepo.GetByID(ctx, *structureID)
	if err != nil {
		return "", "", "", ""
	}
	structName = st.Name
	adminLevel = st.Level
	if st.Level == domain.AdminLevelWoreda {
		woredaName = st.Name
		if st.ParentID != nil {
			parent, errP := s.structureRepo.GetByID(ctx, *st.ParentID)
			if errP == nil {
				subCityName = parent.Name
			}
		}
	} else if st.Level == domain.AdminLevelSubCity || st.Level == "SUBCITY" {
		subCityName = st.Name
	}
	return
}

// GetUserByID loads user info
func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*dto.UserDTO, error) {
	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	structName, adminLevel, subCity, woreda := s.resolveJurisdictionDetails(ctx, u.StructureID)

	return &dto.UserDTO{
		ID:            u.ID,
		FullName:      u.FullName,
		Email:         u.Email,
		PhoneNumber:   u.PhoneNumber,
		Role:          u.Role,
		StructureID:   u.StructureID,
		StructureName: structName,
		AdminLevel:    adminLevel,
		SubCityName:   subCity,
		WoredaName:    woreda,
	}, nil
}

// CreateUser provisions a new official user account with strict jurisdiction validation
func (s *AuthService) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*dto.UserDTO, error) {
	existing, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// 1. የስራ ድርሻ ከስልጣን ወሰን (Jurisdiction) ጋር መጣጣሙን ማረጋገጥ (Strict RBAC Scoping)
	if req.Role == domain.RoleWoredaOfficer {
		if req.StructureID == nil {
			return nil, errors.New("ለ WOREDA_OFFICER ክፍለ ከተማ እና ወረዳ መመረጥ አለበት! (Sub-City and Woreda must be selected for WOREDA_OFFICER)")
		}
		st, err := s.structureRepo.GetByID(ctx, *req.StructureID)
		if err != nil || st.Level != domain.AdminLevelWoreda {
			return nil, errors.New("ለ WOREDA_OFFICER ክፍለ ከተማ እና ወረዳ መመረጥ አለበት! (WOREDA_OFFICER must be assigned to a specific Woreda)")
		}
	} else if req.Role == domain.RoleSubcityManager {
		if req.StructureID == nil {
			return nil, errors.New("ለ SUBCITY_MANAGER ክፍለ ከተማ መመረጥ አለበት! (Sub-City must be selected for SUBCITY_MANAGER)")
		}
		st, err := s.structureRepo.GetByID(ctx, *req.StructureID)
		if err != nil || (st.Level != domain.AdminLevelSubCity && st.Level != "SUBCITY") {
			return nil, errors.New("ለ SUBCITY_MANAGER ክፍለ ከተማ መመረጥ አለበት! (SUBCITY_MANAGER must be assigned to a Sub-City level)")
		}
	}

	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	u := &domain.User{
		ID:           uuid.New(),
		FullName:     req.FullName,
		Email:        req.Email,
		PhoneNumber:  req.PhoneNumber,
		PasswordHash: hash,
		Role:         req.Role,
		StructureID:  req.StructureID,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return s.GetUserByID(ctx, u.ID)
}

// ListUsers returns all official staff & administrators
func (s *AuthService) ListUsers(ctx context.Context) ([]dto.UserDTO, error) {
	users, err := s.userRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]dto.UserDTO, len(users))
	for i, u := range users {
		structName, adminLevel, subCity, woreda := s.resolveJurisdictionDetails(ctx, u.StructureID)

		res[i] = dto.UserDTO{
			ID:            u.ID,
			FullName:      u.FullName,
			Email:         u.Email,
			PhoneNumber:   u.PhoneNumber,
			Role:          u.Role,
			StructureID:   u.StructureID,
			StructureName: structName,
			AdminLevel:    adminLevel,
			SubCityName:   subCity,
			WoredaName:    woreda,
		}
	}
	return res, nil
}

// DeleteUser removes an official user
func (s *AuthService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.Delete(ctx, id)
}
