# ERP System Design & Tradeoffs

---

## 1. Tenant Isolation

**Architecture: why did you structure it this way?**

I went with shared database, shared schema, `tenant_id` on every table,
enforced by Postgres Row-Level Security, instead of schema-per-tenant
or a fully separate database per tenant. Schema-per-tenant is the
"easier to reason about" option (you basically can't leak data across a
schema boundary by accident), but it also doesn't scale operationally —
every migration has to run N times, and it sidesteps the actual hard
problem instead of solving it. I wanted to solve the harder, more common
real-world version: proving isolation holds _inside_ a shared schema,
which is what most real SaaS products (Salesforce-shape systems)
actually do at scale.

**State management: how do you enforce it, and why that mechanism?**

Every tenant-scoped request runs inside a transaction where I call
`SELECT set_config('app.tenant_id', $1, true)` before touching any
table. RLS policies on every table check
`tenant_id = current_setting('app.tenant_id')::uuid`. The important
detail is _how_ I set that value, I deliberately used `set_config` as
a parameterized function call instead of string-building a
`SET LOCAL app.tenant_id = '<value>'` statement, because Postgres
doesn't allow bound parameters inside `SET`, and building that string
by hand is a SQL-injection-shaped habit even when the value is "just a
UUID."

**Failure handling: what happens when this is bypassed or misused?**

I have exactly two Postgres roles: `app_user` (RLS-bound, used for
literally everything) and `app_service` (`BYPASSRLS`, used _only_ for
the handful of operations that structurally cannot know a tenant yet —
signup's first insert, login's cross-tenant email lookup, and a payment
webhook's `checkout_reference` correlation). All three of those are the
same underlying problem: you can't scope a query by tenant when finding
out _which_ tenant is the entire point of the query. Everything else
goes through `app_user`. The failure mode I'm actually worried about
isn't "RLS has a bug", Postgres's RLS implementation is mature and
well-tested, it's "a future code path accidentally uses the service
pool when it didn't need to," which would silently disable isolation
for that one query. That's a discipline problem, not a technology
problem, and I don't currently have anything automated catching it.

**Idempotency / edge cases considered:**

The one deliberately asymmetric policy is on `tenants` itself: `SELECT`
is RLS-scoped normally, but `INSERT` uses a permissive `WITH CHECK
(true)`, because signup is the one legitimate case of creating a tenant
row before that tenant's own context could possibly exist yet.

**What I'd do differently with more time:**

An automated test that iterates every table in the schema and asserts
it (a) has `tenant_id`, (b) has RLS enabled, and (c) has a policy —
failing CI if a new table ever gets added without one. Right now that
check is "did I remember," and I'd rather it be "the build fails if I
forgot."

---

## 2. Cross-Module Consistency (the Event Bus)

**Architecture: why an in-process bus instead of Kafka?**

This is a modular monolith, not distributed services, there's no
network hop between "Sales publishes an event" and "Inventory reacts to
it," so a message broker's actual value proposition (durability across
process boundaries, replay after a consumer restarts, decoupled
scaling) doesn't apply yet. An in-process `Publish`/`Subscribe` gets me
the same _code-level_ decoupling, Sales and Inventory never import
each other, only a shared `events` package, without the operational
cost of running and reasoning about Kafka for a workload that doesn't
need it yet.

**State management: how do you track what happened, and why?**

The genuinely interesting decision here isn't the bus itself, it's that
**publisher and subscriber transactions are not shared.** Sales commits
an order in its own transaction; Inventory's stock-decrement handler
opens a completely separate transaction, on a separate connection, to
react to that same event. I could have tried to thread a shared
`pgx.Tx` through the event so everything commits or rolls back
together, but that would mean Inventory can never manage its own
database session independently of whoever happens to be publishing an
event at it, which stops making sense the moment any of this is ever
split into real services.

**Failure handling: what's the actual consequence of that choice?**

Because the transactions aren't shared, an order can commit
successfully and _then_ fail to be fulfilled (insufficient stock). I
don't hide that, `sales.Service.PlaceOrder` explicitly marks the order
`stock_issue` as a compensating action when the `order.placed` publish
fails, rather than leaving it silently `placed` with stock that was
never actually decremented. The exact same shape happens on the
Procurement side: receiving a PO commits independently of whether
Inventory's restock succeeds, and a failed restock marks the PO
`receive_issue`. If the compensating write _itself_ fails, I surface
both errors together (`fulfillment failed (%w) AND marking order as
stock_issue failed (%w)`) instead of one silently swallowing the other.

**Idempotency / edge cases considered:**

`Publish` runs every subscriber for an event type even if an earlier
one returned an error, Finance still generates an invoice for an order
regardless of whether Inventory's decrement succeeded, because those
are two genuinely independent business concerns and I didn't want one
module's failure to silently suppress another's. I also have a nested
publish (Inventory's `order.placed` handler can itself publish
`stock.low` mid-handler) safe here because the chain provably
terminates, but it's a real thing to watch before a fifth event type
ever gets added.

**What I'd do differently with more time:**

An outbox pattern: write the intent to publish an event as part of the
_same_ transaction as the triggering write, then have a separate
process actually deliver it with retries. Right now, if a subscriber
handler fails for a transient reason (a dropped connection, not a real
business failure), there's no retry, it's marked as a compensating
state and that's the end of it. A retry-with-backoff layer between
"the event fired" and "the handler ran" is the natural next step before
this pattern would hold up under real production failure rates.

---

## 3. Concurrency & Atomicity

**Storage design: what did you choose, and why?**

Two places in this system need "read a number, decide if an action is
allowed, then change the number" under concurrent access: usage caps
(`usage_counters`) and stock quantities (`stock_levels`). Both use the
identical pattern, not a `SELECT` followed by an `UPDATE` in
application code, but a single atomic statement:

```sql
UPDATE stock_levels SET quantity = quantity + $1
WHERE ... AND quantity + $1 >= 0
RETURNING quantity;
```

If zero rows come back, the action was disallowed (would go negative /
over cap) that's a normal outcome the caller checks for, not an
error path.

**Concurrency: how do you handle concurrent access, and what's the
trade-off?**

The naive version read the current quantity in Go, compare it to the
delta, then issue an `UPDATE` has a real race: two concurrent
decrements can both read a quantity that looks safe to subtract from
before either write actually lands, and stock goes negative under load.
Folding the check into the `WHERE` clause of the write itself means
Postgres's own row-level locking serializes concurrent writers against
that row, the database does the mutual exclusion for me instead of me
trying to build it in the application layer. The trade-off is
throughput on a single hot row: if one product is being decremented by
many concurrent orders, those writes serialize against each other by
design. That's the correct trade-off for correctness, but it's a real
ceiling.

**Failure handling / eviction (the equivalent concept here):**

There isn't a TTL-based eviction concept in this system the way Traced
had one, the closer analogy is the monthly reset of `usage_counters`,
which is handled by keying rows on `(tenant_id, period_start)` rather
than mutating a single row in place. A new calendar month is simply a
new row (lazily created via `INSERT ... ON CONFLICT DO NOTHING` on
first use that period), so there's no explicit eviction/cleanup step
needed, old rows just stop being queried, though nothing currently
prunes them, which is worth flagging as its own small gap.

**What breaks first under 10x load:**

The single hot-row lock contention described above, a viral product
with everyone ordering it simultaneously would serialize on that one
`stock_levels` row and become a real bottleneck before almost anything
else in the system does. Second candidate: the `app_user` connection
pool under a spike of concurrent `db.WithTenant` transactions, since
every one of them holds a connection for the duration of its
transaction.

**What I'd do differently with more time:**

For the stock-contention case specifically: sharding a very high-demand
product's stock across multiple counter rows (classic "hot key"
mitigation, the same idea as sharding a popular leaderboard entry) and
summing them for reads, trading a slightly more complex read path for
much better write concurrency on the few SKUs that actually need it.

---

## 4. Payment Orchestration

**Architecture: why did you structure it this way?**

I originally designed this as one `Charge()` method with routing and
failover on every call. That doesn't match how Paystack or Flutterwave
actually work anymore, neither allows charging a raw card server-side
(PCI compliance), so the real shape has to split into
`InitializeCheckout` (a hosted redirect, the only way to get a first
payment) and `ChargeWithToken` (a direct recurring charge using a token
obtained from that first payment). I corrected the design once I hit
that constraint rather than forcing the original shape to fit, worth
being upfront that this was a correction, not the original plan working
out.

**State management: how do you track payment state, and why?**

A webhook arrives with **zero tenant context**, it's an anonymous
HTTP POST from Paystack or Flutterwave's infrastructure, not an
authenticated request from one of my own users. I correlate it back to
a specific tenant's subscription via `checkout_reference`, a value I
generate, send to the gateway as their `reference`/`tx_ref`, and get
echoed back verbatim in the webhook payload. Looking that up runs on
the `BYPASSRLS` service role, for the same structural reason login's
email lookup does: you can't scope a tenant-isolated query by a tenant
you don't know yet.

I deliberately do **not** change a subscription's plan or status while
a checkout is pending, only the transient `checkout_reference` marker
gets set. If I'd modeled "pending" as a subscription status change,
`GetActiveSubscription` (which entitlement checks depend on) would stop
finding an active subscription for the entire pending window, which
would break ordinary product/order creation for a tenant who's simply
in the middle of paying. Keeping them on their current plan until a
webhook actually confirms success avoided that entirely.

**Failure handling: what's your retry strategy? How do you handle
partial failures?**

Failover exists **only** for the first checkout, and only for genuine
gateway-side failures (timeout, 5xx, malformed response) never for a
card decline, which both adapters translate into a distinct
`ErrCardDeclined` and return as-is. Retrying a declined card on the
other gateway wouldn't help (the card is the problem, not the gateway)
and would be actively wrong to attempt silently. Recurring charges
never fail over at all, a Paystack authorization code is meaningless
to Flutterwave, so once a subscription's token belongs to one gateway,
that gateway is where its renewals live permanently.

**Idempotency: how did you implement it? What edge cases did you
consider?**

`payment_transactions.gateway_ref` is `UNIQUE`. Both gateways document
that webhook delivery can happen more than once for the same event, so
a duplicate delivery hitting that constraint is a no-op update, not a
second charge recorded or a subscription activated twice.

**What you'd do differently with more time / in production:**

Two concrete gaps I already know about rather than pretending they
don't exist: recurring charges currently send an empty `email` field to
`ChargeWithToken`, because nothing stores a tenant's billing email for
reuse, it's only ever supplied fresh at checkout time. And a failed
renewal currently just flips a subscription to `past_due` with no grace
period or retry schedule, real dunning logic (a few retry attempts
over some days, an email warning, then downgrade) is the honest next
step before this would be production-ready.

---

## 5. Caching

**Architecture: why cache these specific things, and not others?**

I didn't reach for "add Redis in front of everything", I went through
each read endpoint and asked whether it's read often, written rarely,
and safe to serve slightly stale. `billing.ListPlans` (global,
essentially never changes), and per-tenant lists for warehouses,
products, customers, and vendors (all low write frequency) qualified.
`inventory.GetStock`, `billing.GetUsage`, and `procurement.
ListPurchaseOrders` deliberately did not, each for a different reason —
which is the more interesting design story than "what did you cache."

**State management: how do you track what's cached, and why that key
scheme?**

Every tenant-scoped cache key goes through one function,
`cache.TenantKey(tenantID, ...)`, never a hand-built string. This
matters specifically _because_ this is a multi-tenant system: Postgres
RLS automatically backstops a forgotten `WHERE tenant_id` clause; Redis
has no equivalent concept of a tenant at all, a cache key is just a
string. Key-construction discipline is the _entire_ enforcement
mechanism for tenant isolation at the cache layer, and centralizing key
construction into one function is what makes that discipline
enforceable rather than hoped-for.

**Failure handling: what's your strategy, and how do you handle partial
failures?**

Every cache operation degrades to "treat as a miss" or a silent no-op,
logged but never returned as an error the caller has to react to. If
Redis is down, slow, or simply not configured, every endpoint still
returns correct data from Postgres, just without the speedup. I built
it this way deliberately: a cache is an optimization, and an
optimization that can make a request fail is a liability, not an
optimization.

**Idempotency / invalidation: how did you implement it? What edge cases
did you consider?**

Two mechanisms, doing different jobs on purpose: a short TTL (5
minutes) as a safety net for the case an invalidation gets missed (a
bug, a direct SQL write bypassing the app), and explicit `cache.Delete`
immediately after every corresponding write commits, as the _primary_
mechanism, the one that should be doing almost all the actual
invalidation work in normal operation. `billing.ListPlans` is the one
TTL-only exception, and it's TTL-only for a specific reason, not
laziness: there's no write endpoint that changes plans at all, so
explicit invalidation has nothing to hook into.

The single most important judgment call in this whole section:
`inventory.GetStock` is **not cached**, on purpose, because
`AdjustStock`'s entire design is a single atomic conditional `UPDATE`
built specifically to prevent a race under concurrent writers. A cache
sitting in front of that read would be the exact same race,
reintroduced, the cache becomes a second, staler source of truth for
the one number whose correctness the atomic update exists to guarantee.

**What you'd do differently with more time / in production:**

Cache warming on deploy (pre-populate the plans cache rather than
taking the first request's latency hit), and if any list endpoint ever
needs pagination, the key scheme would need to include the page/offset, right now every cached list is "the whole list," which works at this
scale but wouldn't survive a tenant with thousands of products.

---

## 6. What Breaks First (System-Wide)

Not a single component's failure mode, where does _this whole system_
actually hit a wall first if load grows 10x?

**First to break: the recurring billing worker.**
`RunRecurringBillingOnce` does one unpaginated sweep, `ListDueSubscriptions`
pulls every tenant across the entire platform whose subscription is due,
in one query, then charges them one at a time in a single goroutine
loop. At current scale that's fine. At 10x tenant count, this becomes a
single-threaded bottleneck with no backpressure and no batching, it's
the one piece of the system I built without really load-testing the
assumption "this list stays small."

**Second: the hot-row stock contention described in Section 3.** A
single viral product being ordered concurrently by many tenants'
customers would serialize writes against that one row before most
other parts of the system show strain.

**Third: RLS policy evaluation cost at very high row counts.** RLS
policies add a `WHERE` clause to effectively every query, evaluated per
row, cheap at the row counts this system has today, but worth
knowing it's not free, and worth actually measuring before assuming it
stays negligible at 100x the data.

**What I'd do differently:**
Paginate and batch the billing worker before anything else on this
list, it's the cheapest fix with the clearest failure mode, unlike the
hot-row problem (needs a real sharding decision) or the RLS cost
question (needs to actually be measured with real data volume before
deciding whether it's worth addressing at all).
