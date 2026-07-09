package tenancy

import (
	"context"
	"errors"
	"fmt"

	"github.com/Oliveszn/OneDesk/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrEmailTaken = errors.New("email already registered")

type Repository struct {
	db *db.DB
}

func NewRepository(d *db.DB) *Repository {
	return &Repository{db: d}
}

func (r *Repository) CreateTenantWithAdmin(ctx context.Context, name, email, passwordHash string) (*Tenant, *User, error) {
	var tenant Tenant
	var user User

	err := pgx.BeginFunc(ctx, r.db.Service, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx,
			`INSERT INTO tenants (name) VALUES ($1) RETURNING tenant_id, name, created_at`,
			name,
		).Scan(&tenant.TenantID, &tenant.Name, &tenant.CreatedAt)
		if err != nil {
			return fmt.Errorf("inserting tenant: %w", err)
		}

		if _, err := tx.Exec(ctx, `SELECT set_config('app.tenant_id', $1, true)`, tenant.TenantID.String()); err != nil {
			return fmt.Errorf("setting tenant context: %w", err)
		}

		err = tx.QueryRow(ctx,
			`INSERT INTO users (tenant_id, email, password_hash, role)
			 VALUES ($1, $2, $3, 'admin')
			 RETURNING user_id, tenant_id, email, role, created_at`,
			tenant.TenantID, email, passwordHash,
		).Scan(&user.UserID, &user.TenantID, &user.Email, &user.Role, &user.CreatedAt)
		if err != nil {
			if isUniqueViolation(err) {
				return ErrEmailTaken
			}
			return fmt.Errorf("inserting admin user: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return &tenant, &user, nil
}

// GetUserByEmail is the login lookup, we dont know which tenant this email belongs to untill after this query returns
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.db.Service.QueryRow(ctx,
		`SELECT user_id, tenant_id, email, password_hash, role, created_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&u.UserID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // caller treats "not found" as invalid credentials
		}
		return nil, fmt.Errorf("looking up user by email: %w", err)
	}
	return &u, nil
}

// GetTenant is meant to be called inside a transaction
func (r *Repository) GetTenant(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID) (*Tenant, error) {
	var t Tenant
	err := tx.QueryRow(ctx,
		`SELECT tenant_id, name, created_at FROM tenants WHERE tenant_id = $1`,
		tenantID,
	).Scan(&t.TenantID, &t.Name, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("getting tenant: %w", err)
	}
	return &t, nil
}

func isUniqueViolation(err error) bool {
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23505"
	}
	return false
}
