# Payout Orchestrator — Project Scope

| | |
| --- | --- |
| **Status** | Draft v0.5 (open for review) |
| **Author** | Gabriel Roriz |
| **Updated** | 2026-11-27 |
| **See also** | [Business](00-business.md) · [Design](02-design.md) |

## 1. Running example

Who the customer is and what problem they have is in the [business definition](00-business.md). This document states what the system must do.

*"A creator earns $2 per 1,000 views on a post, up to $500 per post."* It looks trivial, but it isn't:

| What happens | What the system must do |
| --- | --- |
| A view snapshot arrives twice (a client retry) | Count it once: ingestion is idempotent |
| A snapshot of 9k views arrives after one of 12k | Snapshots are cumulative: keep the highest value, so the late one changes nothing |
| The post is removed as fraud after its earnings were paid | Post a negative adjustment (a clawback). Never edit history |
| A snapshot arrives 2 days after the post's 30-day window closed | Accept it within the grace period, ignore it after |
| The rate changes to $1.50 per 1,000 views mid-month | Version the rules, and record which rule version each earning used |
| 12,345 views at $2 per 1,000 is $24.69 | Apply an explicit rounding policy: pay per whole 1,000 views ($24.00) or pro rata ($24.69). No floats anywhere near money |
| The provider times out while paying the creator | Find out whether the payment happened first. A blind retry pays twice |

## 2. Goals and non-goals

### Goals

- **G1.** Turn post events and metrics into creator earnings through declarative, versioned rules.
- **G2.** Keep a double-entry ledger as the single source of truth for every cent.
- **G3.** Route payouts across several providers, with failover that cannot create duplicate payments.
- **G4.** Give tenants a clean API and signed webhooks. Give operators traceability and replay.

### Non-goals and constraints

**Non-goals (v1):** KYC/AML and tax reporting (integrated, not built), holding funds or moving money (licensed providers do that), reading social platforms or tracking metrics (the tenant sends them through the API), FX pricing (provider quotes are used as-is), creator-facing dashboards, and segments other than creator platforms (see [possible future segments](00-business.md#12-possible-future-segments)).

**Constraints:** Go, event-driven over Apache Kafka, AWS. The full stack must also run locally with one command.

## 3. Glossary

- **Tenant:** a creator or influencer platform that uses Payout Orchestrator.
- **Payee:** a creator (a person or their company) the tenant pays. Has payout methods and a verification status.
- **Performance event:** an immutable fact about a post (`post.published`, `post.metrics_snapshot`, `post.removed`) with a client-supplied unique `event_id`.
- **Program:** a named creator incentive, such as pay per view, with one or more immutable **rule versions**. Payees join it through an **enrollment**.
- **Earning:** money a payee is owed. It may be negative (an adjustment).
- **Funding account:** the tenant's single balance, held with a licensed partner, that tops up each provider. We record its balance; we never hold the money.
- **Payout:** an instruction to move money from a payee's balance to one of their payout methods.
- **Route:** the ordered list of providers eligible for a given payout.

## 4. Functional requirements (MVP)

| ID | Requirement |
| --- | --- |
| FR-01 | Every resource is scoped to a tenant. API keys are per tenant and environment (`test` / `live`). |
| FR-02 | Create payees and their payout methods. Sensitive details are encrypted or tokenized. |
| FR-03 | A payout is blocked unless the payee is `verified` (by an external KYC provider or manually). |
| FR-04 | Tenants create programs from fixed [rule building blocks](02-design.md#building-blocks): event filter, measure (latest value of a metric per post, or a count of posts), window (N days after publishing), threshold or tiers, reward (`per_unit`, `fixed` per post) and required caps (per post, per creator per cycle, and a program budget). |
| FR-05 | A rule version is immutable once active and applies from its `effective_from` date. Never retroactively. |
| FR-06 | Payees are enrolled into programs explicitly, optionally restricted by an eligibility filter. |
| FR-07 | Tenants can post **direct earnings** for logic the building blocks can't express. |
| FR-08 | Ingest events over REST, singly or in batches of up to 1,000. Duplicate `event_id`s are ignored. |
| FR-09 | A `post.removed` event drops the post's reward to zero. If the post's earnings were already accrued, it produces a negative earning. |
| FR-10 | Every earning carries its lineage: the contributing events and the rule version. |
| FR-11 | Every money movement is a balanced double-entry journal entry. Corrections are new entries, never edits. |
| FR-12 | Payee balances move `pending → available → in_flight → paid`. Earnings become `available` when their window and grace period close. A negative balance is netted against future earnings. |
| FR-13 | The ledger records the tenant's funding balance: deposits increase it, paid payouts decrease it. A payout the balance can't cover is `blocked`, is never sent to a provider, and the tenant is notified. Blocked payouts resume automatically once funds arrive. |
| FR-14 | Create payouts on demand, for all payees with an `available` balance or for chosen ones. Every mutating call takes an `Idempotency-Key`. |
| FR-15 | Payouts follow an explicit [state machine](02-design.md#4-payout-state-machine). Every transition is stored with its actor and reason. |
| FR-16 | Providers sit behind one adapter interface. The MVP ships a fault-injecting simulator and one real sandbox adapter. |
| FR-17 | Per-tenant routing policies match country, currency and method to an ordered provider list. |
| FR-18 | Fail over on a definitive rejection or an open circuit breaker. **Never** on an ambiguous outcome until it is resolved, and never on a lack of funds. |
| FR-19 | Track provider status through webhooks, with polling for payouts stuck past an SLA. |
| FR-20 | Send tenants HMAC-signed webhooks for earnings and payouts, retried with backoff. |

**Later (v1+):** campaign-level programs, rule backtests, late events as adjustments, scheduled payout cycles per program (for example weekly, with a minimum amount), batch approval, payout cancellation, a second provider, provider rate limits and batch APIs, daily reconciliation, and ops tooling for replay and manual resolution.

## 5. Non-functional requirements

| ID | Quality | Requirement |
| --- | --- | --- |
| NFR-01 | Correctness | Zero duplicate payouts, and zero lost or double-counted earnings. When in doubt, stop and flag. |
| NFR-02 | Ledger integrity | Every journal entry sums to zero per currency. Every balance can be rebuilt from the entries. |
| NFR-03 | Durability | No acknowledged event, earning or ledger entry is ever lost. Ingest keeps accepting events while downstream is down (≥ 7-day buffer). |
| NFR-04 | Performance | Ingest p99 < 100 ms. Event → earning p99 < 5 s. 5k events/s sustained. |
| NFR-05 | Ordering and isolation | Per-payee ordering from ingest to ledger. One tenant's batch can't delay another's payouts. |
| NFR-06 | Auditability | Full lineage: payout → ledger entries → earnings → events → rule version. |
| NFR-07 | Determinism | The same events and rule version always produce the same earnings. |

## 6. Open questions

1. Which provider failure modes hurt most in practice? The answer decides which NFRs come first.
2. Can a removed post be clawed back after its window has closed, or only inside it?
