package service

import (
	"context"
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

	structName := ""
	var adminLevel domain.AdminLevel
	if u.StructureID != nil {
		st, err := s.structureRepo.GetByID(ctx, *u.StructureID)
		if err == nil {
			structName = st.Name
			adminLevel = st.Level
		}
	}

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
		},
		ExpiresAt: expiresAt,
	}, nil
}

// GetUserByID loads user info
func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*dto.UserDTO, error) {
	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	structName := ""
	var adminLevel domain.AdminLevel
	if u.StructureID != nil {
		st, err := s.structureRepo.GetByID(ctx, *u.StructureID)
		if err == nil {
			structName = st.Name
			adminLevel = st.Level
		}
	}

	return &dto.UserDTO{
		ID:            u.ID,
		FullName:      u.FullName,
		Email:         u.Email,
		PhoneNumber:   u.PhoneNumber,
		Role:          u.Role,
		StructureID:   u.StructureID,
		StructureName: structName,
		AdminLevel:    adminLevel,
	}, nil
}

// CreateUser provisions a new official user account
func (s *AuthService) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*dto.UserDTO, error) {
	existing, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
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
		structName := ""
		var adminLevel domain.AdminLevel
		if u.Structure != nil {
			structName = u.Structure.Name
			adminLevel = u.Structure.Level
		} else if u.StructureID != nil {
			st, err := s.structureRepo.GetByID(ctx, *u.StructureID)
			if err == nil {
				structName = st.Name
				adminLevel = st.Level
			}
		}

		res[i] = dto.UserDTO{
			ID:            u.ID,
			FullName:      u.FullName,
			Email:         u.Email,
			PhoneNumber:   u.PhoneNumber,
			Role:          u.Role,
			StructureID:   u.StructureID,
			StructureName: structName,
			AdminLevel:    adminLevel,
		}
	}
	return res, nil
}

// DeleteUser removes an official user
func (s *AuthService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.Delete(ctx, id)
}
