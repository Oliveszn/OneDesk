package tenancy

import (
	"context"
	"fmt"

	"github.com/Oliveszn/OneDesk/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	repo *Repository
	db   *db.DB
}

func NewService(repo *Repository, d *db.DB) *Service {
	return &Service{repo: repo, db: d}
}

func (s *Service) GetMyTenant(ctx context.Context, tenantID uuid.UUID) (*TenantResponse, error) {
	var resp TenantResponse

	err := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		t, err := s.repo.GetTenant(ctx, tx, tenantID)
		if err != nil {
			return fmt.Errorf("repo get tenant: %w", err)
		}

		resp = TenantResponse{
			TenantID:  t.TenantID.String(),
			Name:      t.Name,
			CreatedAt: t.CreatedAt,
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("tenant transaction failed: %w", err)
	}

	return &resp, nil
}
