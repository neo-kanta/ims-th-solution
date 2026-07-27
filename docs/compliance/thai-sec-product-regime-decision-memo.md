# Thai SEC Product-Regime Decision Memo

Status: draft for compliance-expert/owner decision — NOT approved
Program: `IMS-REG-TH-SEC`
Prepared: 2026-07-21 (manager session, merge-blocker remediation program)
Feeds from: `docs/compliance/thai-sec-regulatory-source-map.md` (verified 2026-07-15)

This memo does not select a regime and does not authorize any enforcement
code. It exists so a compliance expert can make exactly one decision: which
product regime IMS implements first. `regulatory.thai_sec` remains a
non-enforcing `NOT_CONFIGURED` stub until that decision and its
source-to-rule matrix are both approved.

## Why this decision is blocking

IMS currently has no field that distinguishes a retail mutual fund from an
accredited-investor (AI) fund, institutional/ultra-high-net-worth (UI) fund,
private fund, or provident fund. Each regime uses a different Thai SEC
appendix with different ratio limits, eligible-investor rules, and disclosure
requirements. Implementing any threshold before this decision would mean
inventing legal requirements — explicitly forbidden by the source map and by
`docs/MANAGER/MEMORY.md`'s Regulatory Program Decisions section.

## Candidate regimes

| Profile | What it is | Included | Excluded | Governing schedule |
| --- | --- | --- | --- | --- |
| Retail mutual fund (recommended starting point) | Publicly offered collective investment scheme | General retail investors subscribing via a licensed distributor | Accredited/institutional-only offers, privately mandated portfolios | Tor Nor. 87/2558 (as amended, currently Tor Nor. 1/2569) retail-MF appendix |
| Accredited-investor fund (AI) | Fund restricted to accredited investors under Thai SEC suitability rules | Investors meeting the AI net-worth/income/experience test | Retail public offers | AI appendix, same base notification family |
| Institutional/UHNW fund (UI) | Fund restricted to institutional or ultra-high-net-worth investors | Institutional investors, UHNW individuals meeting the UI threshold | Retail, most AI-only investors | UI appendix |
| Private fund (PF) | Discretionary/advisory mandate for a single or small group of clients | Clients under a bespoke investment management agreement | Pooled public/AI/UI fund investors | PF appendix plus the client's approved mandate |
| Provident fund (PVD) | Employee retirement savings fund | Fund members under an employer scheme | All other investor types | PVD appendix plus the fund committee's approved policy |

## Proposed first regime

**Recommendation: retail mutual fund**, carried over unchanged from the
2026-07-15 source-map session's own recommendation, conditioned on the owner
confirming it matches the real licensed business. `AIREAD.md` describes IMS
as "a Proof of Concept for an asset management company in Thailand" — this
does not by itself establish which regime(s) the real business is licensed
for. The retail-MF appendix is also the most complete and best-documented
Thai SEC schedule, and IMS's current domain concepts (`Fund`, `LIVE`
portfolio, existing asset-class/concentration/cash rule families already
built in `compliance/rules`) map most directly onto retail-MF-style
diversification and cash-buffer controls.

**This recommendation is not a decision.** The compliance expert/owner must
confirm it against the actual license(s) held, because implementing the wrong
regime's ratios would be a compliance failure, not a technical one.

## Products and customers included/excluded (retail-MF candidate)

- Included: units of a publicly registered mutual fund distributed to retail
  investors through a licensed selling agent; ordinary subscription/
  redemption; NAV-based pricing.
- Excluded from this first regime: any AI/UI-only offer, any private/
  provident-fund mandate, and any portfolio marked outside the retail-MF
  legal wrapper even if it uses the same `LIVE` portfolio type in IMS today.
  IMS's `LIVE`/`SIMULATION`/`MODEL` portfolio types are an operational
  distinction, not a legal product-regime field — see the Known Strategic
  Gaps note below.

## Authoritative Thai SEC sources

Carried from `docs/compliance/thai-sec-regulatory-source-map.md` (do not
duplicate maintenance of this list — that file is the source of truth if the
two ever disagree):

1. Thai SEC Rule Book, Capital Market Supervisory Board Notification Tor Nor.
   87/2558, *Investment of Funds*, effective consolidated text and appendices.
2. The retail mutual fund appendix specifically (ratio schedule), once the
   regime is confirmed.
3. Thai SEC liquidity-risk management rules and the mutual-fund regulations
   index for related retail-MF requirements (disclosure, redemption).
4. Sor Nor. 92/2558 for derivative/embedded-derivative exposure, if the first
   rule set includes derivatives.
5. The fund's own SEC-approved scheme/prospectus, when stricter than the
   appendix minimum or when it supplies fund-specific policy.

## Effective dates and amendment history

- As of 2026-07-15 (still current at this memo's writing, unverified since),
  the operative consolidated notification is **Tor Nor. 1/2569** (35th
  amendment to Tor Nor. 87/2558), signed 2026-02-09, effective 2026-03-01.
- Every enforced rule must record the exact notification number, appendix,
  clause/row, and effective period it implements — not just "Thai SEC" as a
  generic citation. This must be re-verified at implementation time in case a
  newer amendment has since taken effect; the SEC rulebook page linked in the
  source map is the authoritative current-text source, not this memo.

## Source-to-rule mapping (draft skeleton — not approved)

This repeats the source map's draft matrix filtered to the retail-MF
candidate. Every cell is unapproved until the compliance expert confirms the
exact clause and numeric limit:

| Control family | Retail-MF authority | IMS data required | Status |
| --- | --- | --- | --- |
| Eligible asset / asset class | Retail-MF Appendix 3 | Instrument asset classification (already modeled; see `compliance/rules/allocation`) | Not approved |
| Single issuer / group concentration | Appendix 4 + calculation Appendix 5 | Issuer/group reference data (partially modeled; see `compliance/rules/concentration`) | Not approved |
| Sector exposure | Fund scheme/prospectus unless retail-MF sets a floor | Versioned sector taxonomy (modeled; see `compliance/rules/ratio`) | Not approved |
| Liquidity | Thai SEC liquidity-risk rules | Liquidity classification, redemption profile — **not currently modeled in IMS** | Not approved |
| Restricted instruments/issuers | Retail-MF law + approved restricted-list policy | Versioned restricted list — **not currently modeled** | Not approved |
| NAV/AUM and cash controls | Retail-MF appendix + fund scheme | Valuation snapshot, cash ledger (modeled; existing min-cash-buffer rule) | Not approved |
| Derivatives | Sor Nor. 92/2558, if in scope for the fund's scheme | Exposure method, counterparty, collateral — **not currently modeled** | Not approved |
| Credit quality | Retail-MF appendix + credit-rating notification | Rating source/scale/date — stub exists (`credit_rating.minimum`), disabled in the UI catalog | Not approved |

Rows marked "not currently modeled" mean building that control requires new
domain fields before any threshold can be enforced, independent of the legal
approval gate.

## Ambiguities requiring legal/compliance interpretation

- Whether the real licensed business currently holds a retail-MF license, an
  AI/UI license, or operates only private/provident mandates — this memo
  cannot answer that from source code or seed data, and must not guess from
  fund names or demo seeds (explicit prohibition in `MEMORY.md`).
- Whether a single IMS deployment must support more than one regime
  concurrently (e.g., retail-MF funds plus a separate private-fund mandate)
  and, if so, how that is recorded per fund/portfolio without making fund
  association optional.
- Whether liquidity, restricted-list, and derivative controls are required
  for the first enforceable rule set or can be deferred to a later increment
  once the regime is confirmed.
- Whether an existing `SIMULATION`/`MODEL` portfolio should ever be subject to
  the same regime's thresholds (current `MEMORY.md` policy already exempts
  them from the existing asset-classification fail-closed rule; the Thai SEC
  regime decision should state explicitly whether that carries over).
- Legal permissibility and process for an override/exception path (the
  existing IMS compliance engine has generic PASS/WARN/BLOCK plus override
  approval; the compliance expert must confirm whether Thai SEC rules permit
  any override at all, and under what documented justification).

## Validation examples (illustrative only — not approved thresholds)

These are structural examples of what a source-to-rule matrix entry must
contain once approved; the numbers are placeholders and must not be used as
real limits:

- Example row shape: "Single issuer concentration — Appendix 4, clause X —
  maximum Y% of NAV per issuer, excluding government securities — effective
  2026-03-01 under Tor Nor. 1/2569 — BLOCK on breach, no override permitted."
- Example boundary test once approved: a portfolio at exactly the limit
  percentage should PASS; one basis point over should BLOCK; a breach caused
  solely by market movement (passive breach) should follow whatever
  grace-period/remediation rule the compliance expert approves, not an
  immediate forced sale.

## Approval owner

The Thai SEC product-regime decision and its source-to-rule matrix require
sign-off from Kanta (repository owner) or a designated compliance expert
before any implementation begins. Per `docs/MANAGER/MEMORY.md`, the required
approvals are:

1. The first product profile (this memo recommends retail mutual fund,
   pending confirmation).
2. The exact source-to-rule matrix with notification/appendix/clause,
   effective dates, calculation basis, PASS/WARN/BLOCK behavior, override
   permissibility, and evidence/audit fields for every control family in
   scope for the first increment.

Until both approvals are recorded, `regulatory.thai_sec` must keep returning
`NOT_CONFIGURED` rather than a false PASS or an invented BLOCK, and no
threshold, severity, or override policy may be implemented from this memo or
any other document that lacks that sign-off.
