CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- this block creates two db roles 
-- app_user is a standard db user meant for normal app operations
-- app_service is a powerful user that can bypass RLS and see all data across all tenants usually meant for analytics
-- change the change me part to better passwrod
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'app_user') THEN
    CREATE ROLE app_user LOGIN PASSWORD 'changeme_app_user';
  END IF;
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'app_service') THEN
    CREATE ROLE app_service LOGIN PASSWORD 'changeme_app_service' BYPASSRLS;
  END IF;
END
$$;

CREATE TABLE IF NOT EXISTS tenants (
  tenant_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name       TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users (
  user_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id     UUID NOT NULL REFERENCES tenants(tenant_id),
  email         TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  role          TEXT NOT NULL DEFAULT 'admin', -- admin, manager, staff
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (email)
);

-- This gives both of your new database users the permission to read, write, update, and delete data on the tenants and users tables
GRANT SELECT, INSERT, UPDATE, DELETE ON tenants, users TO app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON tenants, users TO app_service;

-- important part: says dont let app_user see anything in these tables unless they pass policy rules below
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE users   ENABLE ROW LEVEL SECURITY;

-- A tenant can only see their own tenant record if the database session's custom setting app.tenant_id matches their ID
CREATE POLICY tenant_self_select ON tenants
  FOR SELECT
  USING (tenant_id = current_setting('app.tenant_id', true)::uuid);

-- anyone can insert a new tenant, for signups
CREATE POLICY tenant_signup_insert ON tenants
  FOR INSERT
  WITH CHECK (true);

-- A user can only see or modify rows in the users table if the row's tenant_id matches the current session's app.tenant_id
CREATE POLICY tenant_isolation ON users
  USING (tenant_id = current_setting('app.tenant_id', true)::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::uuid);