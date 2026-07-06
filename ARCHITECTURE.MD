# Multi-Tenant ERP SaaS Architecture

## Overview

A multi-tenant Enterprise Resource Planning (ERP) system delivered as SaaS, multiple independent businesses ("tenants") use the same deployment, with strict data isolation between them. The system is built as a **modular monolith** in Go: a single deployable service with clean internal module boundaries (Identity, Inventory, Sales, Finance, Procurement, Billing), communicating through an in-process event bus rather than direct cross-module calls.

Tenant isolation is enforced with a **shared database, shared schema** model, every table carries a `tenant_id` column backed by **Postgres Row-Level Security (RLS)** as the actual security boundary, not just application-level filtering.

The platform itself is monetized with **Free and Paid tiers**, enforced via usage caps, and paid subscriptions are billed automatically on a recurring basis through a **payment orchestration engine** that sits in front of two payment gateways: **Paystack** and **Flutterwave** routing each transaction to the better-suited gateway and failing over to the other if a gateway-side error occurs.

---

## Functional Requirements

- Businesses (tenants) sign up and get an isolated workspace with their own users, data, and roles
- User management with role-based access control (RBAC), scoped per tenant
- **Inventory**: manage products, SKUs, stock levels across warehouses
- **Sales**: manage customers, place orders, track order line items
- **Finance**: generate invoices, record ledger entries, track payments
- **Procurement**: manage vendors, raise purchase orders
- Cross-module business workflows happen automatically:
  - Placing an order decrements stock and generates an invoice
  - Stock falling below a threshold triggers a purchase order suggestion
  - Receiving a purchase order restocks inventory
- **Billing**: each tenant is on a **Free** or **Paid** plan; Free enforces usage caps (max users, products, orders/month), Paid removes them
- Paid tenants are billed **automatically on a recurring cycle** (e.g. monthly) via a stored payment method
- Payments are processed through a **payment orchestration engine** that supports two gateways **Paystack** and **Flutterwave** choosing between them per transaction and retrying on the other if one fails

---

## Non-Functional Requirements

- **Tenant isolation is non-negotiable**: a bug in application code must never leak one tenant's data to another. Isolation is enforced at the database layer, not trusted to be remembered in every query.
- **Modular, not distributed**: one deployable service, but internal boundaries strict enough that any module could be extracted into its own service later without a rewrite.
- **Consistency over raw throughput**: this system prioritizes correctness of cross-module workflows (e.g. an order should never decrement stock without also producing an invoice) over handling extreme scale.
- **Auditability**: inventory and financial changes need a traceable history, this is an ERP; "what happened and when" matters as much as current state.
- **Plan limits must be enforced synchronously**: a tenant on the Free plan must be blocked from exceeding a cap at the moment of the action (e.g. creating the 51st product), not caught later by a background job.
- **Payment operations must be idempotent**: gateway webhooks can be delivered more than once, and network timeouts must never result in a tenant being double-charged.
- **No blind retries on card declines**: gateway failover is for gateway-side failures (timeouts, downtime, malformed responses) a genuine card decline is a customer-side outcome and must not be silently retried on a second gateway.

---

## Scale Assumptions

This is designed to demonstrate correct multi-tenant SaaS architecture, not hyperscale throughput. Assumed target: low hundreds of tenant businesses, each with a handful to a few dozen concurrent users. The interesting engineering problems here are **isolation correctness** and **cross-module consistency**, not request volume, so the design optimizes for those.

---

## Tenant Isolation Strategy

**Approach: shared database, shared schema, `tenant_id` on every table, enforced by Postgres RLS.**

Why this over schema-per-tenant or database-per-tenant: it's the standard pattern for SaaS at this scale (Salesforce-style), it keeps migrations to a single schema instead of fanning them out across N tenant schemas, and it's the more interesting isolation problem to solve correctly, anyone can isolate tenants by giving each one a separate database; the harder and more common real-world case is proving isolation holds inside a shared one.

**How a request gets scoped to a tenant:**

1. User authenticates, receives a JWT with a `tenant_id` claim
2. Auth middleware extracts `tenant_id` from the JWT on every request
3. At the start of each request's database transaction, the middleware runs:
   ```sql
   SET LOCAL app.tenant_id = '<tenant-uuid>';
   ```
4. Every table has an RLS policy like:

   ```sql
   ALTER TABLE products ENABLE ROW LEVEL SECURITY;

   CREATE POLICY tenant_isolation ON products
     USING (tenant_id = current_setting('app.tenant_id')::uuid);
   ```

5. Application code still includes `tenant_id` in queries as defense in depth, but **RLS is the real enforcement layer**, even a repository query that forgets the `WHERE tenant_id = ...` clause cannot return or modify another tenant's rows, because Postgres itself refuses.

This is deliberately a place to "break it and observe": write a query without the tenant filter and confirm RLS still blocks cross-tenant access. That failure-mode test is worth calling out explicitly in this doc and in commit history, it's the whole point of choosing RLS over trusting application discipline.

---

## Architecture Style: Modular Monolith

One Go binary. Each business domain lives in its own internal package with its own tables, and modules **never** query another module's tables directly or reach into another module's internals, they only interact through:

- **Exported service interfaces** for synchronous needs (e.g. Sales asking Inventory "is this in stock?")
- **Events** for asynchronous side effects (e.g. Sales announcing "order placed", without knowing or caring who reacts)

```
erp-saas/
  cmd/
    api/                 → main.go, server bootstrap, wiring
  internal/
    tenancy/             → tenant CRUD, JWT tenant-claim middleware
    auth/                → login, JWT issuing/verification, RBAC
    inventory/           → products, warehouses, stock, event handlers
    sales/                → customers, orders, order items
    finance/             → invoices, ledger entries, payments (tenant-facing, e.g. customer invoices)
    procurement/          → vendors, purchase orders
    billing/               → plans, subscriptions, usage counters, entitlement checks
    payments/               → orchestration engine + gateway adapters (paystack, flutterwave)
    events/               → in-process event bus (publisher/subscriber interfaces)
    db/                   → connection pool, migrations, RLS session setup
  migrations/             → manual SQL migrations
```

Note the distinction: **`finance`** handles a tenant's _own_ invoicing of _their_ customers (part of the ERP itself). **`billing`** + **`payments`** handle the platform charging the _tenant_ for using the SaaS. Same shape of problem, two different directions of money, worth being explicit about this in the doc so it doesn't read as a duplicated module.

This keeps the door open to extracting, say, Finance into its own service later, the interface boundary already exists, only the transport would change (in-process call → network call, in-process event → message broker).

---

## Event Bus: In-Process Pub-Sub

Rather than an external broker, cross-module communication uses an **in-process event bus**, a simple publish/subscribe registry living inside the same Go process:

```go
type Event struct {
    Type     string
    TenantID uuid.UUID
    Payload  any
}

type EventBus interface {
    Publish(ctx context.Context, event Event)
    Subscribe(eventType string, handler func(context.Context, Event) error)
}
```

Handlers are registered at startup (e.g. Inventory subscribes to `order.placed`; Finance subscribes to `order.placed`; Procurement subscribes to `stock.low`). Publishing runs registered handlers, either synchronously within the same request or dispatched to a worker goroutine, depending on whether the workflow needs a guaranteed-ordered result before responding to the client.

**Events in use:**

| Event          | Published by | Consumed by | Effect                           |
| -------------- | ------------ | ----------- | -------------------------------- |
| `order.placed` | Sales        | Inventory   | Decrement stock for ordered SKUs |
| `order.placed` | Sales        | Finance     | Generate invoice for the order   |
| `stock.low`    | Inventory    | Procurement | Suggest/raise a purchase order   |
| `po.received`  | Procurement  | Inventory   | Restock received quantities      |

**Explicit tradeoff (worth its own section in `TRADEOFFS.md`)**: an in-process bus has no durability and no replay, if the process crashes mid-dispatch, an event is lost. This is acceptable here because the monolith is a single process with no network hop to fail independently, and because it avoids Kafka's operational overhead for a workload that doesn't yet need cross-service decoupling. If this were split into microservices, this is exactly the interface (`EventBus`) that would get a Kafka-backed implementation without changing any calling code.

---

## Billing: Free vs Paid Plans

Plans differ purely by **usage caps** — no feature-gating for now, which keeps enforcement to one mechanism instead of two:

| Plan | Max users | Max products | Max orders/month | Price                    |
| ---- | --------- | ------------ | ---------------- | ------------------------ |
| Free | 3         | 50           | 100              | ₦0                       |
| Paid | Unlimited | Unlimited    | Unlimited        | recurring monthly charge |

**Enforcement is synchronous and in-process**, unlike the cross-module workflows above, this can't be an event fired after the fact, because the whole point is to block the action before it happens. Every module that creates a capped resource calls the Billing module directly before writing:

```go
// in sales module, before inserting an order
if err := billing.CheckEntitlement(ctx, tenantID, billing.ResourceOrders); err != nil {
    return ErrPlanLimitReached
}
```

`billing.CheckEntitlement` reads the tenant's current plan and a `usage_counters` row for the current billing period, and compares against the plan's cap. This is a plain synchronous function call, not a bus event, the caller needs the answer immediately, not eventually.

---

## Payment Orchestration Engine

A single interface abstracts over both gateways so the rest of the system never talks to Paystack or Flutterwave directly:

```go
type PaymentGateway interface {
    Name() string
    Charge(ctx context.Context, req ChargeRequest) (ChargeResult, error)
    VerifyWebhook(payload []byte, signature string) (WebhookEvent, error)
}
```

`FlutterwaveGateway` and `PaystackGateway` each implement this interface, wrapping their respective REST APIs and translating their gateway-specific error shapes into a common error type the orchestrator can reason about (`ErrGatewayTimeout`, `ErrGatewayDown`, `ErrCardDeclined`, etc. this distinction is what makes safe failover possible).

**Routing logic:**

```go
func (o *Orchestrator) Charge(ctx context.Context, req ChargeRequest) (ChargeResult, error) {
    primary := o.route(req)          // smart routing: pick gateway by currency/corridor
    result, err := primary.Charge(ctx, req)

    if isGatewayFailure(err) {       // timeout / downtime, NOT a card decline
        secondary := o.other(primary)
        result, err = secondary.Charge(ctx, req)
    }
    return result, err
}
```

- **Smart routing**: e.g. NGN transactions route to Paystack first (deep local rails), other currencies/corridors route to Flutterwave first (broader pan-African/multi-currency reach). This routing table is a config, not a code branch per case, easy to extend as more gateways are added.
- **Failover**: only triggers on `isGatewayFailure`, a genuine decline is returned to the caller as-is and never retried on the other gateway, per the non-functional requirement above.
- Every attempt (including failovers) is logged in `payment_transactions.attempted_gateways` so a failed-then-succeeded charge is fully traceable, useful both for debugging and as a demonstrable resilience story.

**Recurring billing mechanics:**

1. On first successful charge, the gateway returns a **tokenized authorization** (Paystack: authorization code; Flutterwave: card token) stored on the tenant's `subscriptions` row, never raw card data
2. A **billing worker** (a background goroutine on a ticker, checked e.g. hourly) queries subscriptions where `current_period_end <= now()`
3. For each due subscription, it calls `Orchestrator.Charge` using the stored token, no card re-entry needed
4. On success: extend `current_period_end`, log the transaction
5. On failure: mark subscription `past_due`; a grace period allows retry before downgrading the tenant to Free

**Webhooks:** both gateways also push async confirmations (`charge.success`, `charge.failed`) to dedicated endpoints. Each webhook handler verifies the gateway's signature, then upserts `payment_transactions` keyed by `gateway_ref` since `gateway_ref` is unique, a duplicate webhook delivery (which both gateways explicitly warn can happen) is a no-op rather than a double-processed payment.

---

## Data Model

All tables below include a `tenant_id UUID NOT NULL` column and an RLS policy identical in shape to the `products` example above; omitted from each snippet for brevity.

### Identity & Tenancy

```sql
CREATE TABLE tenants (
  tenant_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE users (
  user_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   UUID NOT NULL REFERENCES tenants(tenant_id),
  email       TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  role        TEXT NOT NULL, -- 'admin', 'manager', 'staff'
  created_at  TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE (tenant_id, email)
);
```

### Inventory

```sql
CREATE TABLE warehouses (
  warehouse_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    UUID NOT NULL REFERENCES tenants(tenant_id),
  name         TEXT NOT NULL
);

CREATE TABLE products (
  product_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    UUID NOT NULL REFERENCES tenants(tenant_id),
  sku          TEXT NOT NULL,
  name         TEXT NOT NULL,
  UNIQUE (tenant_id, sku)
);

CREATE TABLE stock_levels (
  product_id    UUID NOT NULL REFERENCES products(product_id),
  warehouse_id  UUID NOT NULL REFERENCES warehouses(warehouse_id),
  tenant_id     UUID NOT NULL REFERENCES tenants(tenant_id),
  quantity      INT NOT NULL DEFAULT 0,
  reorder_point INT NOT NULL DEFAULT 0,
  PRIMARY KEY (product_id, warehouse_id)
);
```

### Sales

```sql
CREATE TABLE customers (
  customer_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   UUID NOT NULL REFERENCES tenants(tenant_id),
  name        TEXT NOT NULL,
  email       TEXT
);

CREATE TABLE orders (
  order_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   UUID NOT NULL REFERENCES tenants(tenant_id),
  customer_id UUID NOT NULL REFERENCES customers(customer_id),
  status      TEXT NOT NULL DEFAULT 'placed', -- placed, fulfilled, cancelled
  created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE order_items (
  order_item_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id     UUID NOT NULL REFERENCES tenants(tenant_id),
  order_id      UUID NOT NULL REFERENCES orders(order_id),
  product_id    UUID NOT NULL REFERENCES products(product_id),
  quantity      INT NOT NULL,
  unit_price    NUMERIC(12,2) NOT NULL
);
```

### Finance

```sql
CREATE TABLE invoices (
  invoice_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   UUID NOT NULL REFERENCES tenants(tenant_id),
  order_id    UUID NOT NULL REFERENCES orders(order_id),
  amount      NUMERIC(12,2) NOT NULL,
  status      TEXT NOT NULL DEFAULT 'unpaid', -- unpaid, paid, overdue
  issued_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE ledger_entries (
  entry_id    BIGSERIAL PRIMARY KEY,
  tenant_id   UUID NOT NULL REFERENCES tenants(tenant_id),
  invoice_id  UUID REFERENCES invoices(invoice_id),
  entry_type  TEXT NOT NULL, -- 'debit', 'credit'
  amount      NUMERIC(12,2) NOT NULL,
  created_at  TIMESTAMPTZ DEFAULT NOW()
);
```

### Procurement

```sql
CREATE TABLE vendors (
  vendor_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   UUID NOT NULL REFERENCES tenants(tenant_id),
  name        TEXT NOT NULL
);

CREATE TABLE purchase_orders (
  po_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   UUID NOT NULL REFERENCES tenants(tenant_id),
  vendor_id   UUID NOT NULL REFERENCES vendors(vendor_id),
  status      TEXT NOT NULL DEFAULT 'suggested', -- suggested, sent, received
  created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE po_items (
  po_item_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   UUID NOT NULL REFERENCES tenants(tenant_id),
  po_id       UUID NOT NULL REFERENCES purchase_orders(po_id),
  product_id  UUID NOT NULL REFERENCES products(product_id),
  quantity    INT NOT NULL
);
```

### Billing & Payments

`plans` is the one table in the system **without** `tenant_id`, plans are platform-wide configuration, not tenant data, so RLS doesn't apply to it.

```sql
CREATE TABLE plans (
  plan_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name          TEXT NOT NULL,          -- 'free', 'paid'
  max_users     INT,                    -- NULL = unlimited
  max_products  INT,
  max_orders_per_month INT,
  price_amount  NUMERIC(12,2),
  price_currency TEXT,
  billing_interval TEXT                 -- 'monthly'
);

CREATE TABLE subscriptions (
  subscription_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID NOT NULL REFERENCES tenants(tenant_id),
  plan_id         UUID NOT NULL REFERENCES plans(plan_id),
  status          TEXT NOT NULL DEFAULT 'active', -- active, past_due, cancelled
  gateway         TEXT,                 -- 'paystack' | 'flutterwave' (holds recurring auth)
  gateway_customer_ref TEXT,
  gateway_auth_token   TEXT,            -- tokenized authorization, never raw card data
  current_period_start TIMESTAMPTZ,
  current_period_end   TIMESTAMPTZ,
  created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE payment_transactions (
  transaction_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID NOT NULL REFERENCES tenants(tenant_id),
  subscription_id UUID REFERENCES subscriptions(subscription_id),
  gateway         TEXT NOT NULL,
  gateway_ref     TEXT NOT NULL UNIQUE, -- dedupes webhook retries
  amount          NUMERIC(12,2) NOT NULL,
  currency        TEXT NOT NULL,
  status          TEXT NOT NULL,        -- pending, success, failed
  attempted_gateways TEXT[],            -- failover trail
  created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE usage_counters (
  tenant_id      UUID NOT NULL REFERENCES tenants(tenant_id),
  period_start   TIMESTAMPTZ NOT NULL,
  orders_count   INT NOT NULL DEFAULT 0,
  products_count INT NOT NULL DEFAULT 0,
  users_count    INT NOT NULL DEFAULT 0,
  PRIMARY KEY (tenant_id, period_start)
);
```

---

## API Design

All endpoints require a JWT with a `tenant_id` claim (except signup/login).

| Method    | Endpoint                          | Module      | Purpose                                                      |
| --------- | --------------------------------- | ----------- | ------------------------------------------------------------ |
| POST      | `/v1/tenants`                     | Identity    | Register a new business (tenant)                             |
| POST      | `/v1/auth/login`                  | Identity    | Authenticate, issue JWT                                      |
| GET/POST  | `/v1/products`                    | Inventory   | List / create products                                       |
| GET/PATCH | `/v1/stock/:productId`            | Inventory   | View / adjust stock levels                                   |
| POST      | `/v1/orders`                      | Sales       | Place an order → publishes `order.placed`                    |
| GET       | `/v1/orders/:id`                  | Sales       | Order status and line items                                  |
| GET       | `/v1/invoices`                    | Finance     | List invoices for tenant                                     |
| PATCH     | `/v1/invoices/:id/pay`            | Finance     | Record a payment                                             |
| GET/POST  | `/v1/vendors`                     | Procurement | Manage vendors                                               |
| POST      | `/v1/purchase-orders/:id/receive` | Procurement | Mark PO received → publishes `po.received`                   |
| GET       | `/v1/billing/plans`               | Billing     | List available plans                                         |
| POST      | `/v1/billing/subscribe`           | Billing     | Upgrade to Paid, capture payment method, create subscription |
| GET       | `/v1/billing/usage`               | Billing     | Tenant's current usage vs plan caps                          |
| POST      | `/webhooks/paystack`              | Payments    | Paystack async payment confirmation                          |
| POST      | `/webhooks/flutterwave`           | Payments    | Flutterwave async payment confirmation                       |

---

## High-Level Architecture

```
                     ┌──────────────────────────────┐
Client (per tenant) ─▶      Go HTTP API (chi)        │
                     │  ┌────────────────────────┐   │
                     │  │  Auth + Tenant Middleware│  │ ── extracts tenant_id from JWT
                     │  │  SET LOCAL app.tenant_id │  │     opens per-request transaction
                     │  └────────────────────────┘   │
                     │            │                   │
                     │  ┌─────────┼──────────────┐    │
                     │  ▼         ▼              ▼     │
                     │ Identity Inventory ... Procurement ── Billing (entitlement checks)
                     │  (internal modules, package-bounded)      │
                     │            │                              │
                     │      In-process EventBus            Payment Orchestrator
                     │   (order.placed, stock.low...)     (routes + fails over)
                     └──────────────┬───────────────────────────┬─┘
                                    ▼                           ▼
                          PostgreSQL (Neon)          ┌───────────┴──────────┐
                     RLS enforced on every table     ▼                      ▼
                     (except `plans`, global)   Paystack API         Flutterwave API
                                                      │                      │
                                                      └──────────┬───────────┘
                                                          Webhooks (async confirmation)
                                                                 ▼
                                                     /webhooks/paystack, /webhooks/flutterwave

                     Billing worker (ticker goroutine): scans subscriptions due for
                     renewal → charges via Payment Orchestrator → extends period or
                     marks past_due
```

---

## Order Placement Flow (Step by Step)

```
1. Client calls POST /v1/orders (JWT carries tenant_id)
2. Middleware sets Postgres session tenant_id for the transaction
3. Sales module validates customer + line items, inserts order + order_items
4. Sales publishes `order.placed` event on the in-process bus
5. Inventory handler (subscribed to order.placed) decrements stock_levels
   → if resulting quantity < reorder_point, Inventory publishes `stock.low`
6. Finance handler (subscribed to order.placed) generates an invoice + ledger entry
7. Procurement handler (subscribed to stock.low) creates a suggested purchase_order
8. Response returned to client once the order transaction commits
```

Steps 5–7 can run synchronously within the request for strong consistency, or be dispatched to background goroutines for lower request latency, this is a deliberate design decision to document and justify in `TRADEOFFS.md` (consistency vs. latency).

---

## Failure Scenarios

| Failure                                                          | Impact                                                                         | Recovery                                                                                                     |
| ---------------------------------------------------------------- | ------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------ |
| App forgets `WHERE tenant_id` in a query                         | None, RLS blocks cross-tenant rows regardless                                  | N/A by design                                                                                                |
| Process crashes mid event-handler dispatch                       | In-process bus event is lost; e.g. invoice might not be generated for an order | No automatic replay (documented tradeoff); would require moving to a durable broker                          |
| Postgres connection pool exhausted under load                    | Writes/reads start failing                                                     | Same batching/pooling lessons from the leaderboard project apply, batch writes, bound pool size              |
| A module's handler panics                                        | Could block other handlers if run synchronously in the same goroutine          | Wrap each handler dispatch in a recover(), log and continue                                                  |
| Primary gateway times out mid-charge                             | Charge attempt fails                                                           | Orchestrator fails over to the secondary gateway; both attempts logged in `attempted_gateways`               |
| Card is genuinely declined                                       | Charge fails                                                                   | **Not** retried on the other gateway, returned to the caller as-is, since this is a customer-side outcome    |
| Webhook delivered twice (both gateways document this can happen) | Risk of double-processing a payment                                            | `gateway_ref UNIQUE` constraint makes the second insert a no-op                                              |
| Billing worker crashes mid-cycle                                 | Some due subscriptions not charged that cycle                                  | Next worker run picks up any subscription still past `current_period_end`; charging is idempotent per period |
| Recurring charge fails for a subscription                        | Tenant would lose paid features immediately                                    | Subscription marked `past_due` with a grace period before downgrade, rather than an instant cutoff           |

---

## Future Considerations

- **Schema-per-tenant migration path** for large enterprise customers who require stronger isolation guarantees than RLS alone (a common SaaS tiering strategy, "your data is in a dedicated schema on request")
- **Swap in-process EventBus for Kafka** if/when modules are extracted into separate services, the `EventBus` interface is designed so this is an implementation swap, not a rewrite
- **Read replicas / tenant-based sharding** if a single Postgres instance becomes a bottleneck at high tenant counts
- **Idempotency keys** on order placement to safely handle client retries without double-decrementing stock
- **Feature-gating on top of usage caps** (e.g. lock Procurement to Paid only) once usage-cap enforcement is proven out
- **Dunning flow**: automated retry schedule + email reminders before downgrading a `past_due` tenant to Free
- **A third gateway** (e.g. Stripe for non-African customers) the `PaymentGateway` interface is designed so this is an additive adapter, not a rewrite of the orchestrator
