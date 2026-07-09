package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/Oliveszn/OneDesk/internal/tenancy"
	"github.com/Oliveszn/OneDesk/internal/token"
	"github.com/Oliveszn/OneDesk/internal/validate"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type Service struct {
	repo  *tenancy.Repository
	token *token.JWTService
}

func NewService(repo *tenancy.Repository, t *token.JWTService) *Service {
	return &Service{repo: repo, token: t}
}

func (s *Service) Signup(ctx context.Context, req SignupRequest) (*AuthResponse, error) {
	if req.BusinessName == "" {
		return nil, errors.New("business_name is required")
	}
	if err := validate.Email(req.Email); err != nil {
		return nil, fmt.Errorf("validate email: %w", err)
	}
	if err := validate.Password(req.Password); err != nil {
		return nil, fmt.Errorf("validate password: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	tenant, user, err := s.repo.CreateTenantWithAdmin(ctx, req.BusinessName, req.Email, string(hash))
	if err != nil {
		return nil, fmt.Errorf("create tenant DB error: %w", err)
	}

	jwt, err := s.token.Issue(user.UserID, tenant.TenantID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("issuing token: %w", err)
	}

	return &AuthResponse{
		Token:    jwt,
		TenantID: tenant.TenantID.String(),
		UserID:   user.UserID.String(),
		Role:     user.Role,
	}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("looking up user: %w", err)
	}
	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return nil, ErrInvalidCredentials
	}

	jwt, err := s.token.Issue(user.UserID, user.TenantID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("issuing token: %w", err)
	}

	return &AuthResponse{
		Token:    jwt,
		TenantID: user.TenantID.String(),
		UserID:   user.UserID.String(),
		Role:     user.Role,
	}, nil
}
