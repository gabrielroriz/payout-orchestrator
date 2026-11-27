# Payout Orchestrator — Business Definition

| | |
| --- | --- |
| **Status** | Draft v0.4 (open for review) |
| **Author** | Gabriel Roriz |
| **Updated** | 2026-11-27 |
| **Next** | [Scope](01-scope.md) · [Design](02-design.md) |

## 1. In one sentence

Payout Orchestrator lets a creator platform pay its creators for views, posts and campaigns, by sending post events and program rules to one API, without building payments infrastructure.

## 2. The customer

Our customer (the **tenant**) is a creator or influencer platform.

| Who gets paid | Paid for |
| --- | --- |
| Creators | Views, posts, campaigns |

Creator platforms pay fewer, larger and more international payouts than most performance programs. That plays to our strengths: multi-provider coverage and failover.

Within the tenant, three groups use the product:

- **Growth / creator partnerships** designs the programs: what is rewarded, and how much.
- **Engineering** integrates the API: it sends post events and metrics, and handles webhooks.
- **Finance / ops** funds the account, triggers payouts and reconciles them against the ledger.

The creators being paid (**payees**) are the tenant's users. They never deal with us directly.

## 3. The problem

Without us, a tenant has to build all of this:

1. **Calculating what creators are owed.** Code that turns views and posts into amounts, and stays correct when metrics arrive duplicated, late or out of order, and when posts are removed for fraud.
2. **Reaching creators everywhere.** Each payout provider covers only some countries, currencies and methods, so a global creator base needs several integrations.
3. **Surviving provider failures.** When a provider is down or times out, payouts stall or risk being sent twice.
4. **Keeping the books.** Finance needs an audit trail from every payment back to the posts and metrics that earned it.

That is months of payments engineering, unrelated to the tenant's product, and expensive to get wrong.

## 4. What we offer

| Tenant need | What we provide |
| --- | --- |
| Launch creator programs quickly | Programs configured through the API, from a set of rule building blocks |
| Pay the right amount | Exact money math, with duplicate and late metrics, and removed posts, handled automatically |
| Pay creators anywhere | One API in front of several payout providers |
| Never stall, never pay twice | Automatic failover between providers, safe against duplicates |
| Satisfy finance and audit | A double-entry ledger, statements (later), and full lineage from payout to event |
| Stay informed | Signed webhooks for earnings and payouts |

What we deliberately leave out is listed in the [scope's non-goals](01-scope.md#2-goals-and-non-goals).

## 5. How it works for a tenant

Example: *ClipCo*, a fictional creator platform, pays creators $2 for every 1,000 views.

1. **ClipCo's growth team designs the program.** The terms are ClipCo's business decision.
2. **ClipCo's engineers create the program** through our API. We validate it against the platform limits; there is no manual approval from us.
3. **A creator joins** from inside ClipCo. ClipCo registers the creator with us as a payee and enrolls them in the program.
4. **ClipCo sends view counts** as events. We apply the rule and record what each creator earned, first as `pending`.
5. **The program's settlement cycle closes** (for example, the week ends). Earnings become `available`, after a grace period for late metrics and removed posts.
6. **ClipCo's finance team funds one account** and pays out by clicking a button. Later, payouts can also run on the program's schedule.
7. **We pay each creator** through the best provider for their country and method, exactly once, and notify ClipCo.

Creators see only ClipCo. We are the engine underneath, the way a payment processor sits under a checkout page.

## 6. What tenants control, and what we control

| The tenant decides | The platform enforces |
| --- | --- |
| Which programs to run, and their terms (thresholds, rates, caps, windows) | Programs are built only from a fixed set of rule building blocks |
| Which creators join each program | Caps and budgets are required, and rate limits apply per tenant |
| When to pay out: on demand, and later on the program's schedule | Payees must be verified, and the funding balance must cover the payout |
| Their routing policy: which providers to use, in what order, per country, currency and method | Payouts go only through providers that support the route |

For logic the building blocks cannot express, tenants can calculate amounts themselves and post them as direct earnings, which then follow the same ledger and payout path.

## 7. Alternatives the tenant has today

| Alternative | Why it falls short |
| --- | --- |
| Build in-house | Months of payments work, plus ongoing maintenance and risk |
| Use one provider's bulk payout API directly | No rules engine or ledger, gaps in coverage, and stuck when that provider fails |
| Spreadsheets and manual payouts | Doesn't scale, error-prone, no audit trail |
| Performance-payroll platforms, such as [Grade](https://www.usegrade.com/) | The category this project belongs to. Our focus is the engine for creator payouts: exact amounts, no duplicate payouts, and no single-provider dependency |

## 8. Funding

The tenant funds **one account**, held with a licensed partner, not with us. We instruct transfers from it to each provider's balance as payouts need them. This means:

- The tenant manages one balance, not one per provider.
- Failover works without the tenant prefunding every provider.
- We never hold funds: we only send instructions.

We do **record** the balance in our ledger, and check it before sending any payout.

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

## 12. Possible future segments

Other companies also pay people for results, and could reuse the same ledger, payouts and routing, with new rule building blocks. They are out of scope for now.

| Segment | Who gets paid | Paid for |
| --- | --- | --- |
| Affiliate and referral programs | Affiliates, referrers | Conversions, sign-ups |
| Gig and delivery marketplaces | Couriers, workers | Deliveries, completed jobs |
| Loyalty and cashback apps | End customers | Purchases, visits, milestones |
| AI agent marketplaces | Agent operators | Completed tasks |

Consumer cashback, with millions of small payouts, would need a different cost and capacity profile than creator payouts.

## 13. Open questions

1. Which licensed partner holds the funding account, and how fast can it top up each provider?
2. What volume tiers and price per payout?
3. Should a program be able to pay each earning as soon as it is available, instead of in cycles?
