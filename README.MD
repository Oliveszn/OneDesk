# Multi-Tenant ERP SaaS

A multi-tenant ERP backend in Go inventory, sales, invoicing,
procurement, and billing for multiple businesses running on one shared
deployment, with tenant isolation enforced at the database layer and a
real payment orchestration engine running two production payment
gateways behind it.

Seven backend phases. No ORM. Plain SQL, Row-Level Security, an
in-process event bus connecting four business domains that never
import each other, and a payment engine that actually handles the
parts most implementations skip routing, failover, webhook
verification, and recurring billing done correctly instead of
simulated.

Every endpoint below works today and is testable via Postman/Swagger
(`api-test-plan.md`).

---

## Payment Orchestration Engine

This is the part worth looking at first.

Two real gateways **Paystack** and **Flutterwave** sitting behind
one interface, with the orchestration logic that actually matters in
production:

- **Currency-based routing** on the first checkout (NGN prefers
  Paystack's local rails, other currencies prefer Flutterwave's
  broader reach), with **automatic failover to the other gateway on a
  genuine gateway failure** timeout, 5xx, malformed response
- **Failover never triggers on a card decline.** A declined card is a
  customer-side outcome, not a gateway problem, and retrying it on a
  different gateway would be actively wrong both adapters classify
  this distinctly and the orchestrator respects it
- **The real hosted-checkout + tokenized-recurring-charge flow** both
  gateways actually require for PCI compliance not a simplified
  "charge a card number" API that neither provider has allowed in
  years
- **Webhook handling with zero tenant context.** A payment webhook is
  an anonymous POST with no JWT, no session, nothing to scope a
  multi-tenant query by, resolved through a correlation reference
  generated at checkout and echoed back by the gateway, verified with
  gateway-specific signature schemes (HMAC-SHA512 for Paystack, a
  dashboard-configured secret hash for Flutterwave, genuinely
  different mechanisms, handled correctly for each)
- **Idempotent by construction**: both gateways document that webhook
  delivery can happen more than once; a unique constraint on the
  gateway reference makes a duplicate delivery a no-op instead of a
  double-processed payment
- **A recurring billing worker** that charges renewals directly against
  whichever gateway processed each subscription's original checkout,
  a recurring token from one gateway is meaningless to the other, so
  renewals never route or fail over, by design

Full write-up of every decision here: including the point where the
original design had to be corrected once the real PCI constraints
became clear — is in `TRADEOFFS.md`, Section 4.

---

## Everything Else

- **Tenant isolation enforced at the database layer.** Postgres
  Row-Level Security on every tenant-scoped table, backed by two
  distinct Postgres roles, one RLS-bound for normal requests, one
  `BYPASSRLS` for the handful of operations that structurally can't
  know a tenant yet (login, signup, payment webhooks)
- **An actual event-driven architecture, in one process.** Sales,
  Inventory, Finance, and Procurement communicate exclusively through
  an in-process pub/sub bus. None of them import each other. Delete any
  one subscription line in `main.go` and the two modules it connected
  still compile, proof the boundary is real, not documented
- **Concurrency correctness proven under load, not assumed.** Usage-cap
  enforcement and stock-quantity adjustment both use an atomic
  conditional `UPDATE` specifically because the naive
  read-then-write version has a real race under concurrent requests,
  same pattern, applied independently in two different modules
- **Cross-module failure handled honestly.** When a side effect fails
  after its trigger already committed, an order succeeds but can't be
  fulfilled, the system marks an explicit compensating state
  (`stock_issue`, `receive_issue`) instead of pretending the whole
  operation was atomic when it wasn't
- **Caching with actual judgment**, not Redis bolted onto every
  endpoint. Tenant-scoped cache keys as the sole defense against
  cross-tenant leakage (Redis has no RLS-equivalent), and specific,
  reasoned calls about what's deliberately _not_ cached, stock levels,
  usage counters, anywhere a stale read would be wrong, not just
  slightly outdated

---

## Tech Stack

- **Go**: standard library `net/http` + `chi` for routing
- **PostgreSQL** (via [Neon](https://neon.tech)): `pgx/v5`, Row-Level
  Security, manual SQL migrations, no ORM
- **Redis**: cache-aside via `redis/go-redis`
- **JWT** (`golang-jwt/jwt`)
- **Paystack** and **Flutterwave**: real integrations, not mocks

---

## Architecture

Modular monolith, package-by-feature: every domain owns its own model,
DTO, repository, service, and handler. Shared infrastructure (`events`,
`cache`, `db`, `token`, `middleware`) has zero dependencies back into
any domain.

```
internal/
  tenancy/      Tenant, User
  auth/         signup/login, JWT
  billing/      Plans, Subscriptions, usage entitlements, checkout + webhooks
  payments/     Paystack + Flutterwave adapters, orchestrator (routing + failover)
  inventory/    Warehouses, Products, Stock — publishes AND subscribes to events
  sales/        Customers, Orders — publishes events
  finance/      Invoices, Ledger — subscribes to events only
  procurement/  Vendors, Purchase Orders — subscribes AND publishes events
  events/       in-process pub/sub bus + shared event contracts
  cache/        Redis-backed cache-aside, tenant-scoped key discipline
  db/           connection pools (two Postgres roles), RLS transaction helper
  token/        JWT issuing/parsing
  middleware/   auth middleware
  reqctx/       shared request-context helpers
  httputil/     response helpers
  validate/     input validation
cmd/api/
  main.go       all wiring
  router.go     all routes
```

**Full system design**: event chain diagrams, the complete data model,
every design correction that happened mid-build is in
[`ARCHITECTURE.md`](./ARCHITECTURE.md).

**Every deliberate trade-off**, argued out with the alternative
considered and rejected, is in [`TRADEOFFS.md`](./TRADEOFFS.md).
Caching gets its own deeper write-up in [`CACHING.md`](./CACHING.md).

---

## Getting Started

### 1. Database

Run every migration in `migrations/` **in order**:

```
0001_identity_tenancy.sql
0002_billing_foundations.sql
0003_inventory.sql
0004_sales.sql
0005_finance.sql
0006_procurement.sql
0007_payments.sql
```

`0001` creates the two Postgres roles (`app_user`, `app_service`)
everything else depends on, **change the placeholder passwords**
before running it against anything real.

### 2. Environment

Copy `.env.example` to `.env`:

```bash
APP_DB_DSN=postgres://app_user:<password>@<neon-host>/<db>?sslmode=require
SERVICE_DB_DSN=postgres://app_service:<password>@<neon-host>/<db>?sslmode=require
JWT_SECRET=<long random string>

# Optional: checkout/webhooks won't work without these, nothing else needs them
PAYSTACK_SECRET_KEY=
FLUTTERWAVE_SECRET_KEY=
FLUTTERWAVE_WEBHOOK_HASH=
FLUTTERWAVE_REDIRECT_URL=

# Optional: runs fine without Redis, just without the caching speedup
REDIS_ADDR=
REDIS_PASSWORD=
REDIS_DB=0
```

### 3. Run

```bash
go mod tidy
go run ./cmd/api
```

Listens on `:8080`.

---

## Trying the API

```bash
# Sign up: creates a tenant, first admin user, and a Free plan subscription
curl -X POST localhost:8080/v1/tenants \
  -H 'Content-Type: application/json' \
  -d '{"business_name":"Acme Co","email":"you@example.com","password":"supersecret"}'

curl localhost:8080/v1/tenants/me -H 'Authorization: Bearer <token>'
```

Full endpoint reference: `ARCHITECTURE.md`'s API table. A systematic
route-by-route test plan, happy paths, validation errors,
tenant-isolation checks, concurrency tests is in
[`api-test-plan.md`](./api-test-plan.md).

---

## What's Built

1. **Identity & Tenancy**: signup/login, JWT, RLS wiring
2. **Billing**: Free/Paid plans, atomic usage-cap enforcement
3. **Inventory**: warehouses, products, stock, atomic concurrency-safe adjustments
4. **Sales**: customers, orders, the first cross-module event
5. **Finance**: invoices generated automatically, fully decoupled from Sales
6. **Procurement**: vendors, purchase orders, auto-suggested reorders, closing the event loop
7. **Payment Orchestration**: Paystack + Flutterwave, checkout, webhooks, recurring billing

Plus caching, added once all seven phases existed to reason about.

## Known Gaps

Documented, not hidden, full list with reasoning in
`ARCHITECTURE.md`. No double-pay guard on invoices, no dunning/retry
logic beyond a single `past_due` flip, no batching yet on the recurring
billing worker's tenant sweep.
