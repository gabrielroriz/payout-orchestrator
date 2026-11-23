# Payout Orchestrator — Business Definition

| | |
| --- | --- |
| **Status** | Draft v0.2 (open for review) |
| **Author** | Gabriel Roriz |
| **Updated** | 2026-11-23 |
| **Next** | [Scope](01-scope.md) · [Design](02-design.md) |

## 1. In one sentence

Payout Orchestrator lets a company pay many people for their results, by sending performance events and rules to one API, without building payments infrastructure.

## 2. The customer

Our customer (the **tenant**) is a company that pays many people based on results rather than hours.

| Segment | Who gets paid | Paid for |
| --- | --- | --- |
| Creator and influencer platforms | Creators | Views, posts, campaigns |
| Affiliate and referral programs | Affiliates, referrers | Conversions, sign-ups |
| Gig and delivery marketplaces | Couriers, workers | Deliveries, completed jobs |
| Loyalty and cashback apps | End customers | Purchases, visits, milestones |
| AI agent marketplaces | Agent operators | Completed tasks |

**First target: creator platforms.** They pay fewer, larger and more international payouts. That plays to our strengths: multi-provider coverage and failover. Consumer cashback, with millions of small payouts, comes later.

Within the tenant, three groups use the product:

- **Marketing / growth** designs the programs: what is rewarded, and how much.
- **Engineering** integrates the API: it sends events and handles webhooks.
- **Finance / ops** funds the account, triggers payouts and reads statements.

The people being paid (**payees**) are the tenant's users. They never deal with us directly.

## 3. The problem

Without us, a tenant has to build all of this:

1. **Calculating what people are owed.** Code that turns activity into amounts, and stays correct when events arrive duplicated, late or refunded.
2. **Reaching people everywhere.** Each payout provider covers only some countries, currencies and methods, so a global program needs several integrations.
3. **Surviving provider failures.** When a provider is down or times out, payouts stall or risk being sent twice.
4. **Keeping the books.** Finance needs an audit trail from every payment back to the activity that earned it.

That is months of payments engineering, unrelated to the tenant's product, and expensive to get wrong.

## 4. What we offer

| Tenant need | What we provide |
| --- | --- |
| Launch incentive programs quickly | Programs configured through the API, from a set of rule building blocks |
| Pay the right amount | Exact money math, with duplicates, late events and refunds handled automatically |
| Pay people anywhere | One API in front of several payout providers |
| Never stall, never pay twice | Automatic failover between providers, safe against duplicates |
| Satisfy finance and audit | A double-entry ledger, statements, and full lineage from payout to event |
| Stay informed | Signed webhooks for earnings and payouts |

We **do not** hold funds or move money ourselves (licensed providers do, on our instructions), run KYC or tax reporting (these plug in), or track performance (the tenant sends us events).

## 5. How it works for a tenant

Example: *ClipCo*, a fictional creator platform, pays creators $2 for every 1,000 views.

1. **ClipCo's growth team designs the program.** The terms are ClipCo's business decision.
2. **ClipCo's engineers create the program** through our API. We validate it against the platform limits; there is no manual approval from us.
3. **A creator joins** from inside ClipCo. ClipCo registers the creator with us as a payee and enrolls them in the program.
4. **ClipCo sends view counts** as events. We apply the rule and record what each creator earned, first as `pending`.
5. **The program's window closes** (for example, the week ends). Earnings become `available`, after a grace period for late events and refunds.
6. **ClipCo's finance team funds one account** and pays out, by clicking a button or on the program's schedule.
7. **We pay each creator** through the best provider for their country and method, exactly once, and notify ClipCo.

Creators see only ClipCo. We are the engine underneath, the way a payment processor sits under a checkout page.

## 6. What tenants control, and what we control

| The tenant decides | The platform enforces |
| --- | --- |
| Which programs to run, and their terms (thresholds, rates, caps, windows) | Programs are built only from a fixed set of rule building blocks |
| Which of their users join each program | Caps and budgets are required, and usage limits apply per tenant |
| When to pay out: on demand, or on the program's schedule | Payees must be verified, and the funding balance must cover the payout |
| Their routing preferences (cost or speed) | Payouts go only through providers that support the route |

For logic the building blocks cannot express, tenants can calculate amounts themselves and post them as direct earnings, which then follow the same ledger and payout path.

## 7. Alternatives the tenant has today

| Alternative | Why it falls short |
| --- | --- |
| Build in-house | Months of payments work, plus ongoing maintenance and risk |
| Use one provider's bulk payout API directly | No rules engine or ledger, gaps in coverage, and stuck when that provider fails |
| Spreadsheets and manual payouts | Doesn't scale, error-prone, no audit trail |
| Performance-payroll platforms, such as [Grade](https://www.usegrade.com/) | The category this project belongs to. Our focus is the engine: exact amounts, no duplicate payouts, and no single-provider dependency |

## 8. Funding

The tenant funds **one account**, held with a licensed partner, not with us. We instruct transfers from it to each provider's balance as payouts need them. This means:

- The tenant manages one balance, not one per provider.
- Failover works without the tenant prefunding every provider.
- We never hold funds: we only send instructions.

We do **record** the balance in our ledger. It goes up when a deposit is confirmed and down when a payout is paid. Before sending any payout, we check that the balance covers it.

**Example.** ClipCo deposits $50,000, but its 3,000 creators earned $62,000 this week after a viral video.

| | What happens |
| --- | --- |
| Without a recorded balance | All payouts are sent. About 600 fail at the provider for lack of funds, fail again after failover, and creators get "payment failed" emails. Finance sorts out by hand who was paid. |
| With a recorded balance | $50,000 is paid. The other $12,000 is marked *blocked: insufficient funds*, nothing is sent, and ClipCo gets a webhook. Once ClipCo deposits more, the blocked payouts go out automatically. |

## 9. When payouts happen

- **On demand (MVP):** the tenant calls the API, for example from a "Pay creators" button in its own dashboard. It pays what is `available`, for all payees or for chosen ones.
- **On the program's schedule (later):** the program sets a cycle, such as "every Friday, if the creator has at least $20".

The program sets the default schedule, and the tenant can always pay on demand on top of it.

## 10. Business model

- **A fee per successful payout**, tiered by volume.
- Failed or canceled payouts are free.
- Provider fees and FX are passed through to the tenant.

## 11. Success measures

| Measure | Target (hypothesis) |
| --- | --- |
| Time for a tenant to launch a first program | Days, not months |
| Duplicate payouts | Zero |
| Payouts stalled by a single provider outage | Zero |
| Payouts that need manual reconciliation | < 0.1% |

## 12. Decisions

| Topic | Decision |
| --- | --- |
| First segment | Creator platforms |
| Funding | One funding account per tenant, feeding each provider |
| Funding balance | Recorded in our ledger and checked before every payout. Not held by us |
| Payout timing | On demand (MVP), plus a schedule per program (later) |
| Pricing | Per successful payout, tiered by volume |
| Program flexibility | Rule building blocks, plus a direct-earnings API |

## 13. Open questions

1. Which licensed partner holds the funding account, and how fast can it top up each provider?
2. What volume tiers and price per payout?
3. Which creator-platform metrics come first: views, posts or campaigns?
4. Should a program be able to pay each earning as soon as it is available, instead of in cycles?
