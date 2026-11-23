# Payout Orchestrator — Business Definition

| | |
| --- | --- |
| **Status** | Draft v0.1 (first sketch, open for review) |
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
- **Finance / ops** funds the wallet, approves payout batches and reads statements.

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

Example: *FoodApp*, a fictional restaurant app, launches "Visit 10 restaurants in November, get 5% back".

1. **Marketing designs the campaign.** The terms are FoodApp's business decision.
2. **Engineering creates the program** through our API. We validate it against the platform limits; there is no manual approval from us.
3. **A user joins** from inside FoodApp, for example by tapping "Join". FoodApp enrolls that user with us as a payee.
4. **FoodApp sends each visit** as an event. We track progress and work out the cashback.
5. **Finance funds one account** and triggers payouts, or lets a schedule do it.
6. **We pay the user** through the best provider for their country and method, and notify FoodApp.

The user sees only FoodApp. We are the engine underneath, the way a payment processor sits under a checkout page.

## 6. What tenants control, and what we control

| The tenant decides | The platform enforces |
| --- | --- |
| Which programs to run, and their terms (thresholds, rates, caps, windows) | Programs are built only from a fixed set of rule building blocks |
| Which of their users join each program | Caps and budgets are required, and usage limits apply per tenant |
| When and how much to pay out | Payees must be verified, and the wallet must hold enough funds |
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
- We still never hold funds: we only send instructions.

## 9. Business model

- **A fee per successful payout**, tiered by volume.
- Failed or canceled payouts are free.
- Provider fees and FX are passed through to the tenant.

## 10. Success measures

| Measure | Target (hypothesis) |
| --- | --- |
| Time for a tenant to launch a first program | Days, not months |
| Duplicate payouts | Zero |
| Payouts stalled by a single provider outage | Zero |
| Payouts that need manual reconciliation | < 0.1% |

## 11. Decisions

| Topic | Decision |
| --- | --- |
| First segment | Creator platforms |
| Funding | One funding account per tenant, feeding each provider |
| Pricing | Per successful payout, tiered by volume |
| Program flexibility | Rule building blocks, plus a direct-earnings API |

## 12. Open questions

1. Which licensed partner holds the funding account, and how fast can it top up each provider?
2. What volume tiers and price per payout?
3. Which creator-platform metrics come first: views, posts or campaigns?
