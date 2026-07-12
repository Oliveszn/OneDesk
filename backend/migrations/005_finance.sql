-- This is the tenant's OWN invoicing of THEIR customers

CREATE TABLE IF NOT EXISTS invoices (
  invoice_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id  UUID NOT NULL REFERENCES tenants(tenant_id),
  order_id   UUID NOT NULL REFERENCES orders(order_id),
  amount     NUMERIC(12,2) NOT NULL,
  status     TEXT NOT NULL DEFAULT 'unpaid', -- unpaid, paid, overdue
  issued_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- invoice_id is nullable  not every ledger entry ties back to an invoice 
CREATE TABLE IF NOT EXISTS ledger_entries (
  entry_id   BIGSERIAL PRIMARY KEY,
  tenant_id  UUID NOT NULL REFERENCES tenants(tenant_id),
  invoice_id UUID REFERENCES invoices(invoice_id),
  entry_type TEXT NOT NULL, -- 'debit' (invoice issued) or 'credit' (invoice paid)
  amount     NUMERIC(12,2) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

GRANT SELECT, INSERT, UPDATE, DELETE ON invoices, ledger_entries TO app_user, app_service;
GRANT USAGE, SELECT ON SEQUENCE ledger_entries_entry_id_seq TO app_user, app_service;

ALTER TABLE invoices       ENABLE ROW LEVEL SECURITY;
ALTER TABLE ledger_entries ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON invoices
  USING (tenant_id = current_setting('app.tenant_id', true)::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY tenant_isolation ON ledger_entries
  USING (tenant_id = current_setting('app.tenant_id', true)::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::uuid);