# Dashboard Pattern Language

Read this before building any new module. It documents the reusable
shapes already built — a new module should reuse these components, not
reinvent them. If a new screen needs something not listed here, that's
a sign either a new shared pattern is needed, or the existing ones
should be reconsidered before adding a one-off.

## Layout

**`DashboardShell`** (`dashboard/layout/`) — the only top-level layout.
Sidebar (240px, fixed, ledger-navy `#24344a`) + a content `<Outlet />`.
Every module route renders inside this via React Router nesting under
`/dashboard`. Don't build a new layout wrapper per module.

**`Sidebar`** — one nav item per business module, color-coded to match
the landing page's `Modules.jsx` (Inventory = `#24344a`, Sales =
`#b8863a`, Finance = `#5b7a63`, Procurement = `#16191c`). A module not
yet built shows a "Soon" tag and isn't clickable — see `Sidebar.jsx`'s
`navItems` array; flip `enabled: true` there when a module is ready,
that's the only change needed to "launch" a new nav item.

**`PageHeader`** (`dashboard/shared/`) — every top-level module screen
starts with this: title, optional description, right-aligned action
slot for the page's one primary button (`New Product`, `New Order`,
etc.). No eyebrow label like the landing page uses — inside the
product the sidebar already says where you are.

## Data display

**Tables** — Ant Design `Table`, `pagination={false}` for now (every
current dataset is small; add real pagination once a list can
plausibly exceed a screen). Columns should mirror the actual API
response fields, not an idealized version of them — if a field isn't
in the DTO, it doesn't belong in the table.

**`EmptyState`** (`dashboard/shared/`) — every table's zero-rows case.
Icon + a description of what's actually empty (never generic "no
data") + an optional action. Don't write a bespoke empty state per
table.

**Status tags** — flat `Tag` components, colored from the existing
palette only (`#9c4a3c` danger, `#b8863a` warning/low). Never introduce
a new hue for a new status; find the closest existing meaning first.

## Actions & mutations

**Modals, not separate pages, for create/edit forms** — matches the
landing page's Login/Register pattern. Ant Design `Modal` + `Form`,
`destroyOnClose` so the form resets between opens. A modal's fields
should be exactly the request body fields the corresponding endpoint
accepts — see `NewProductModal.jsx` vs. `CreateProductRequest` in the
backend for the reference example of a 1:1 match.

**Modals are "dumb"** — they collect input and call an `onCreate` /
`onAdjust` prop, they never own the actual data. The parent page (e.g.
`InventoryPage`) owns state and decides what happens on submit. This
is deliberate: when the API gets wired up, only the parent's callback
changes (from "push a mock row" to "POST and refetch"), the modal
itself doesn't change at all.

**Only design actions the backend actually supports.** `ProductsTable`
has no Edit/Delete columns because the backend has no
`PATCH`/`DELETE /v1/products` — see its own comment. Before adding a
button, check the corresponding Go handler exists. A button with
nothing to call is a worse handoff than no button.

## Billing awareness

**`PlanLimitBanner`** (`dashboard/shared/`) — any screen for a capped
resource (see `billing.Resource` in the backend: orders, products,
users) should show this once the tenant is near or at their cap. Takes
`used`/`cap` — map directly to fields on `billing.UsageResponse`. Skips
rendering entirely on an unlimited plan (`cap == null`). See
`InventoryPage.jsx` for the reference wiring.

## Simulating backend constraints in mock data

Where a backend endpoint has a real, meaningful failure mode — not
just "the network could be down," but a _business rule_ like "can't
adjust stock below zero" — the mock version should simulate that
rejection client-side, so the error state exists in the design before
anyone's touched a real API. See `AdjustStockModal.jsx`'s negative-
quantity check as the reference example. A design that only shows the
happy path isn't actually finished.

## Typography & color

Same tokens as the landing page (`src/index.css`'s `@theme` block) —
no new hex values without a documented reason (see `--color-danger`'s
comment for the one addition made so far, and why). Dashboard body
text runs smaller and tighter than the landing page (14px base vs.
16-17px) — this is a working tool, not a pitch.
