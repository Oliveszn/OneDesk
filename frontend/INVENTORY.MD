# Module Spec: Inventory

Read `PATTERNS.md` first — this doc only covers what's specific to
Inventory, not the shared conventions every module reuses.

## Screens

**`/dashboard/inventory`** — `InventoryPage.jsx`. One screen, two tabs
(`Products`, `Warehouses`), not two separate routes — they're small
enough and closely related enough that route-switching would be more
friction than value. Reconsider this if either list grows a lot of
its own screen-specific chrome later (filters, bulk actions).

### Products tab

- Table: SKU (mono), Name, Stock (see below), Actions
- Primary action (top right): **New product** — disabled once the
  tenant is at their product cap (`PlanLimitBanner` + the button's
  `disabled` prop both react to the same `atProductCap` check —
  don't let one without the other, a banner saying "you're at your
  limit" next to an enabled button that will just fail is worse than
  either alone)
- Stock column shows: total quantity across all warehouses (summed,
  not per-warehouse — that detail lives in the Adjust Stock modal),
  plus an **Out of stock** tag (total = 0) or **Low stock** tag (any
  warehouse ≤ its `reorder_point`) — see `ProductsTable.jsx`'s
  `StockSummary`. These two tags are mutually exclusive per the
  current logic (out implies not "low", it's worse) — that's
  deliberate, don't show both on one row.
- Row action: **Adjust stock** only. No edit, no delete — see
  `PATTERNS.md`'s "only design actions the backend supports."

### Warehouses tab

- Table: Name only — that's the entirety of `WarehouseResponse`
  right now. Don't invent columns (address, capacity, etc.) that
  aren't real fields; if the backend grows those fields later, the
  table grows with it, not before.
- Primary action: **New warehouse**

## Modals

**New Product** (`NewProductModal.jsx`) — `sku`, `name`. Maps 1:1 to
`CreateProductRequest`.

**New Warehouse** (`NewWarehouseModal.jsx`) — `name`. Maps 1:1 to
`CreateWarehouseRequest`.

**Adjust Stock** (`AdjustStockModal.jsx`) — the one with real
complexity, worth reading in full before building anything similar:

- Shows current quantity **per warehouse** as read-only context above
  the form (a small bordered list, not a nested table) — the person
  adjusting stock needs to see what they're adjusting _from_ before
  picking a number
- Form fields: `warehouse_id` (select), `delta` (number, explicitly
  signed — the field's `extra` text says so directly rather than
  relying on the person to guess that negative numbers are valid)
- **Client-side simulation of the backend's real rejection**: if
  `current_quantity + delta < 0`, the modal shows an inline `Alert`
  with the same shape of message the real 409 would need to
  communicate (what would happen, why it's blocked) — not a generic
  "invalid input." This is the one place in the whole module where
  getting the error state right matters more than the happy path,
  because it's not a validation nicety, it's the actual business rule
  the backend's atomic `UPDATE` exists to enforce (see
  `inventory_repository.go`'s `AdjustStock` if you want the backend
  side of this same constraint)

## Data contracts (exact shapes, from the real backend DTOs)

```
ProductResponse    { product_id, sku, name }
WarehouseResponse  { warehouse_id, name }
StockLevelResponse { product_id, warehouse_id, quantity, reorder_point }
CreateProductRequest   { sku, name }
CreateWarehouseRequest { name }
AdjustStockRequest     { warehouse_id, delta }
```

`mockData/inventory.js` uses these exact field names. When this gets
wired to the real API: `GET /v1/products` → `products` state,
`GET /v1/warehouses` → `warehouses` state, and
`GET /v1/products/{id}/stock` → populate `stockLevels[productId]`
lazily (the mock version preloads everything up front purely for
convenience; the real version should probably fetch a product's stock
only when its Adjust Stock modal opens, not for every row on page
load).

## Known gaps in this screen (carried over from the backend, or introduced here)

- **No way to set `reorder_point`** — the backend has no endpoint for
  it either (see `ARCHITECTURE.md`'s "Known Gaps" if that file still
  exists in your working copy — it did as of this design pass). The
  mock data hardcodes reasonable values. This needs either a field on
  New Product, an edit affordance somewhere, or its own small modal —
  undecided, worth a real decision before wiring the API rather than
  guessing under time pressure then.
- **No search/filter on either table** — fine at current scale (a
  handful of mock rows), a real problem once a tenant has hundreds of
  products. Not designed yet.
- **Stock summary sums across warehouses with no way to see the
  breakdown without opening Adjust Stock** — that's a real product
  decision (keep the table dense) not an oversight, but worth
  revisiting if "just tell me which warehouse is low" turns out to be
  a common need.
