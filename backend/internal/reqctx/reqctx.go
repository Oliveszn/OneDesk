package reqctx

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey string

const (
	keyUserID   ctxKey = "user_id"
	keyTenantID ctxKey = "tenant_id"
	keyRole     ctxKey = "role"
)

func WithAuth(ctx context.Context, userID, tenantID uuid.UUID, role string) context.Context {
	ctx = context.WithValue(ctx, keyUserID, userID)
	ctx = context.WithValue(ctx, keyTenantID, tenantID)
	ctx = context.WithValue(ctx, keyRole, role)
	return ctx
}

func TenantID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(keyTenantID).(uuid.UUID)
	return id, ok
}

func UserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(keyUserID).(uuid.UUID)
	return id, ok
}

func Role(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(keyRole).(string)
	return role, ok
}
