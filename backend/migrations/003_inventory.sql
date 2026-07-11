CREATE TABLE IF NOT EXISTS warehouses (
  warehouse_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    UUID NOT NULL REFERENCES tenants(tenant_id),
  name         TEXT NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS products (
  product_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id  UUID NOT NULL REFERENCES tenants(tenant_id),
  sku        TEXT NOT NULL,
  name       TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (tenant_id, sku)
);

CREATE TABLE IF NOT EXISTS stock_levels (
  product_id    UUID NOT NULL REFERENCES products(product_id),
  warehouse_id  UUID NOT NULL REFERENCES warehouses(warehouse_id),
  tenant_id     UUID NOT NULL REFERENCES tenants(tenant_id),
  quantity      INT NOT NULL DEFAULT 0,
  reorder_point INT NOT NULL DEFAULT 0,
  PRIMARY KEY (product_id, warehouse_id)
);

GRANT SELECT, INSERT, UPDATE, DELETE ON warehouses, products, stock_levels TO app_user, app_service;

ALTER TABLE warehouses   ENABLE ROW LEVEL SECURITY;
ALTER TABLE products     ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_levels ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON warehouses
  USING (tenant_id = current_setting('app.tenant_id', true)::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY tenant_isolation ON products
  USING (tenant_id = current_setting('app.tenant_id', true)::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY tenant_isolation ON stock_levels
  USING (tenant_id = current_setting('app.tenant_id', true)::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::uuid);