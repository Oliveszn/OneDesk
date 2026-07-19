-- checkout_reference is what ties an async webhook back to the
-- subscription that initiated it, a webhook arrives with no tenant
-- context at all so this is the only thread connecting "a payment succeeded somewhere"
-- to "which tenant's subscription this was for".
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS gateway TEXT;              -- 'paystack' | 'flutterwave', set once the first payment succeeds
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS gateway_auth_token TEXT;   -- tokenized authorization for recurring charges, never raw card data
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS checkout_reference TEXT UNIQUE;

CREATE TABLE IF NOT EXISTS payment_transactions (
  transaction_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id          UUID NOT NULL REFERENCES tenants(tenant_id),
  subscription_id    UUID REFERENCES subscriptions(subscription_id),
  gateway            TEXT NOT NULL,
  gateway_ref        TEXT NOT NULL UNIQUE, -- dedupes duplicate webhook delivery
  amount             NUMERIC(12,2) NOT NULL,
  currency           TEXT NOT NULL,
  status             TEXT NOT NULL, -- pending, success, failed
  attempted_gateways TEXT[],        -- failover trail for the initial checkout only
  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

GRANT SELECT, INSERT, UPDATE, DELETE ON payment_transactions TO app_user, app_service;

ALTER TABLE payment_transactions ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON payment_transactions
  USING (tenant_id = current_setting('app.tenant_id', true)::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::uuid);