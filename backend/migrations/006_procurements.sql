CREATE TABLE IF NOT EXISTS vendors (
  vendor_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id  UUID NOT NULL REFERENCES tenants(tenant_id),
  name       TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS purchase_orders (
  po_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id  UUID NOT NULL REFERENCES tenants(tenant_id),
  vendor_id  UUID REFERENCES vendors(vendor_id),
  status     TEXT NOT NULL DEFAULT 'suggested', -- suggested, sent, received, receive_issue
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS po_items (
  po_item_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    UUID NOT NULL REFERENCES tenants(tenant_id),
  po_id        UUID NOT NULL REFERENCES purchase_orders(po_id),
  product_id   UUID NOT NULL REFERENCES products(product_id),
  warehouse_id UUID NOT NULL REFERENCES warehouses(warehouse_id),
  quantity     INT NOT NULL
);

GRANT SELECT, INSERT, UPDATE, DELETE ON vendors, purchase_orders, po_items TO app_user, app_service;

ALTER TABLE vendors         ENABLE ROW LEVEL SECURITY;
ALTER TABLE purchase_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE po_items        ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON vendors
  USING (tenant_id = current_setting('app.tenant_id', true)::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY tenant_isolation ON purchase_orders
  USING (tenant_id = current_setting('app.tenant_id', true)::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY tenant_isolation ON po_items
  USING (tenant_id = current_setting('app.tenant_id', true)::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::uuid);