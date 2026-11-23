# Payout Orchestrator — Project Scope

| | |
| --- | --- |
| **Status** | Draft v0.4 (open for review) |
| **Author** | Gabriel Roriz |
| **Updated** | 2026-11-23 |
| **See also** | [Business](00-business.md) · [Design](02-design.md) |

## 1. Problem statement

> A platform must work out what each payee earned from raw performance events. It must then pay them through third-party providers exactly once, with a full audit trail, and without depending on any single provider.

Who the customer is and why they need this is in the [business definition](00-business.md).

**Running example:** *"Someone visits 10 restaurants → they earn 5% cashback."* It looks trivial, but it isn't:

| What happens | What the system must do |
| --- | --- |
| Visit #10 arrives twice (a client retry) | Count it once: ingestion is idempotent |
| Visit #4 is refunded after the cashback was paid | Post a negative adjustment (a clawback). Never edit history |
| Visit #10 arrives 2 days after the month closed | Apply a window grace period, or carry the visit into the next period |
| The rule changes to "8 visits → 4%" mid-month | Version the rules, and record which payee is on which version |
| 5% of $123.45 is $6.1725 | Apply an explicit rounding policy. No floats anywhere near money |
| The provider times out while paying the cashback | Find out whether the payment happened first. A blind retry pays twice |

## 2. Goals and non-goals

### Goals

- **G1.** Turn performance events into earnings through declarative, versioned rules.
- **G2.** Keep a double-entry ledger as the single source of truth for every cent.
- **G3.** Route payouts across several providers, with failover that cannot create duplicate payments.
- **G4.** Give tenants a clean API and signed webhooks. Give operators traceability and replay.

### Non-goals and constraints

**Non-goals (v1):** KYC/AML and tax reporting (integrated, not built), holding funds or moving money (licensed providers do that), performance tracking (events arrive through the API), FX pricing (provider quotes are used as-is), and end-user dashboards.

**Constraints:** Go, event-driven over Apache Kafka, AWS. The full stack must also run locally with one command.

## 3. Glossary

- **Tenant:** a company that uses Payout Orchestrator.
- **Payee:** a person or entity the tenant pays. Has payout methods and a verification status.
- **Performance event:** an immutable fact (`visit.completed`) with a client-supplied unique `event_id`.
- **Program:** a named incentive with one or more immutable **rule versions**. Payees join it through an **enrollment**.
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
| FR-04 | Tenants create programs from fixed rule building blocks: event filter, aggregation (`count`, `count_distinct`, `sum`, `max`), window (`calendar_month`, `rolling_Nd`), threshold or tiers, reward (`fixed`, `percent_of_sum`, `per_unit`) and a required cap. |
| FR-05 | A rule version is immutable once active and applies from its `effective_from` date. Never retroactively. |
| FR-06 | Payees are enrolled into programs explicitly, optionally restricted by an eligibility filter. |
| FR-07 | Tenants can post **direct earnings** for logic the building blocks can't express. |
| FR-08 | Ingest events over REST, singly or in batches of up to 1,000. Duplicate `event_id`s are ignored. |
| FR-09 | A reversal event undoes its original. If the original was already accrued, it produces a negative earning. |
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

**Later (v1+):** rule backtests, late events as adjustments, scheduled payout cycles per program (for example weekly, with a minimum amount), batch approval, payout cancellation, a second provider, provider rate limits and batch APIs, daily reconciliation, and ops tooling for replay and manual resolution.

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
2. Does "10 restaurants" mean distinct restaurants (assumed) or 10 visits? 5% of which spend?
