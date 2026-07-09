package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	App     *pgxpool.Pool
	Service *pgxpool.Pool
}

func New(ctx context.Context, appDSN, serviceDSN string) (*DB, error) {
	appPool, err := pgxpool.New(ctx, appDSN)
	if err != nil {
		return nil, fmt.Errorf("connecting app pool: %w", err)
	}

	servicePool, err := pgxpool.New(ctx, serviceDSN)
	if err != nil {
		return nil, fmt.Errorf("connecting service pool: %w", err)
	}

	return &DB{App: appPool, Service: servicePool}, nil
}

func (d *DB) Close() {
	d.App.Close()
	d.Service.Close()
}

// WithTenant opens a transaction on the App pool, sets the Postgres
// session variable app.tenant_id for the lifetime of that transaction
func (d *DB) WithTenant(ctx context.Context, tenantID uuid.UUID, fn func(tx pgx.Tx) error) error {
	tx, err := d.App.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning tx: %w", err)
	}
	defer tx.Rollback(ctx) // no-op if already committed

	if _, err := tx.Exec(ctx, `SELECT set_config('app.tenant_id', $1, true)`, tenantID.String()); err != nil {
		return fmt.Errorf("setting tenant context: %w", err)
	}

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
