CREATE TABLE IF NOT EXISTS plans (
  plan_id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name                 TEXT NOT NULL UNIQUE,   -- 'free', 'paid'
  max_users            INT,                    -- NULL = unlimited
  max_products         INT,
  max_orders_per_month INT,
  price_amount         NUMERIC(12,2),
  price_currency       TEXT,
  billing_interval     TEXT                    -- 'monthly', NULL for free
);

GRANT SELECT ON plans TO app_user, app_service;

INSERT INTO plans (name, max_users, max_products, max_orders_per_month, price_amount, price_currency, billing_interval)
VALUES
  ('free', 3, 50, 100, 0, 'NGN', NULL),
  ('paid', NULL, NULL, NULL, 5000, 'NGN', 'monthly')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS subscriptions (
  subscription_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id            UUID NOT NULL REFERENCES tenants(tenant_id),
  plan_id              UUID NOT NULL REFERENCES plans(plan_id),
  status               TEXT NOT NULL DEFAULT 'active', -- active, past_due, cancelled
  current_period_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  current_period_end   TIMESTAMPTZ,             -- NULL for free (no billing cycle to end)
  created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

GRANT SELECT, INSERT, UPDATE, DELETE ON subscriptions TO app_user, app_service;

ALTER TABLE subscriptions ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON subscriptions
  USING (tenant_id = current_setting('app.tenant_id', true)::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::uuid);

CREATE TABLE IF NOT EXISTS usage_counters (
  tenant_id      UUID NOT NULL REFERENCES tenants(tenant_id),
  period_start   TIMESTAMPTZ NOT NULL,
  orders_count   INT NOT NULL DEFAULT 0,
  products_count INT NOT NULL DEFAULT 0,
  users_count    INT NOT NULL DEFAULT 0,
  PRIMARY KEY (tenant_id, period_start)
);

GRANT SELECT, INSERT, UPDATE, DELETE ON usage_counters TO app_user, app_service;

ALTER TABLE usage_counters ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON usage_counters
  USING (tenant_id = current_setting('app.tenant_id', true)::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::uuid);