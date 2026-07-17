# Thai SEC Regulatory Source Map

Status: draft for owner/compliance approval  
Program: `IMS-REG-TH-SEC`  
Verified: 2026-07-15  
Jurisdiction: Thailand  
Regulator: Securities and Exchange Commission, Thailand (Thai SEC)

## Purpose

This document identifies the authoritative source family for replacing the
non-enforcing `regulatory.thai_sec` warning stub. It is not legal advice and it
does not approve any numeric threshold. Enforcement starts only after the owner
or compliance expert selects the applicable product regime and approves a
source-to-rule matrix.

## Authoritative Source Order

1. The Thai SEC Rule Book for Asset Management and the effective consolidated
   text of Capital Market Supervisory Board Notification Tor Nor. 87/2558,
   *Investment of Funds*, including amendments and currently effective
   appendices:
   https://capital.sec.or.th/webapp/nrs/nrs_main_search.php?SearchNo=&chkNo=1&chkPost=1&doc_id=7080
2. The applicable current appendix for the regulated product profile. The SEC
   rulebook maintains separate ratio schedules for retail mutual funds, AI
   funds, UI funds, private funds, and provident funds.
3. Related Thai SEC notifications and official guidance for the controlled
   subject, including derivatives, credit-rating use, liquidity-risk
   management, NAV, and fund management.
4. The SEC-approved fund scheme/prospectus or the approved private/provident
   fund investment mandate when it is stricter or supplies portfolio-specific
   policy.
5. A versioned internal restricted list or compliance policy approved by the
   responsible compliance/legal owner. Demo seeds and application defaults are
   never authority.

## Currency of the Master Rule

The repository's historical material must not be treated as current law. As of
2026-07-15, the Thai SEC rulebook lists Tor Nor. 1/2569, *Investment of Funds
(35th amendment)*, signed 2026-02-09 and effective 2026-03-01. The implementation
must record the exact notification/appendix version and effective period used
for every enforceable rule.

Useful official entry points:

- Thai SEC rulebook history for Tor Nor. 87/2558:
  https://capital.sec.or.th/webapp/nrs/nrs_main_search.php?SearchNo=&chkNo=1&chkPost=1&doc_id=7080
- Thai SEC mutual-fund regulations index:
  https://www.sec.or.th/TH/Pages/LawandRegulations/MutualFundRegulations.aspx
- Thai SEC mutual-fund management index:
  https://www.sec.or.th/EN/Pages/LAWANDREGULATIONS/MUTUALFUNDMANAGEMENT.aspx
- Thai SEC liquidity-risk management summary:
  https://www.sec.or.th/TH/Pages/LAWANDREGULATIONS/MUTUALFUND-RISKMGT.aspx
- Consolidated derivative/embedded-derivative requirements, Sor Nor. 92/2558:
  https://publish.sec.or.th/nrs/9475p_r.pdf
- Thai SEC private-fund overview and related regulations:
  https://www.sec.or.th/en/pages/lawandregulations/privatefund.aspx
- Thai SEC private/provident fund management summary:
  https://www.sec.or.th/EN/pages/lawandregulations/privatefundmanagementorprovidentfund.aspx

## Required Product-Regime Decision

Select exactly one first enforceable profile:

| Profile | Governing ratio schedule | Current IMS classification |
| --- | --- | --- |
| Retail mutual fund | Retail-MF appendix | Not represented |
| Accredited-investor fund (AI) | AI appendix | Not represented |
| Institutional/ultra-high-net-worth fund (UI) | UI appendix | Not represented |
| Private fund (PF) | PF appendix plus client mandate | Not represented |
| Provident fund (PVD) | PVD appendix plus committee-approved policy | Not represented |

The existing generic `Fund`, portfolio type (`LIVE`, `SIMULATION`, `MODEL`),
fund category, and decision product type do not establish a legal product
regime. A default must not be inferred from the word "fund" or from a demo seed.

## Draft Source-to-Control Matrix

This matrix identifies source families only. Every row remains unapproved until
the applicable clauses, appendix rows, calculations, exceptions, severity, and
effective dates are reviewed.

| Control family | Primary authority | Additional authority/data | Enforcement status |
| --- | --- | --- | --- |
| Eligible asset / asset class | Tor Nor. 87/2558 and effective Appendix 3 | Fund scheme or mandate | Not approved |
| Single issuer / entity / group concentration | Applicable effective Appendix 4 and calculation Appendix 5 | Issuer/group reference data; fund scheme | Not approved |
| Sector exposure | Fund scheme/prospectus or approved mandate unless a product-specific rule applies | Versioned sector taxonomy | Not approved |
| Liquidity | Thai SEC liquidity-risk rules and applicable fund scheme | Liquidity classification, redemption profile, stress data | Not approved |
| Restricted instruments/issuers | Applicable law plus approved restricted-list policy | Versioned list, reason, issuer/instrument identifiers | Not approved |
| NAV/AUM and cash controls | Applicable product rules plus approved scheme/mandate | Official valuation snapshot and cash ledger | Not approved |
| Derivatives / embedded derivatives | Sor Nor. 92/2558 and applicable scheme | Exposure method, counterparty, collateral data | Not approved |
| Credit quality | Applicable appendix and Thai SEC credit-rating notification | Rating source, scale, date, issuer/instrument mapping | Not approved |

## Approval Gate

Before code replaces the warning stub, the owner/compliance expert must approve:

- the first product profile;
- exact notification, clause, appendix row, amendment, and effective dates;
- the calculation basis and required source data;
- PASS/WARN/BLOCK behavior and whether an override is legally permitted;
- exception, grace-period, passive-breach, and remediation behavior;
- evidence fields and source/version metadata retained for audit;
- pre-trade, post-trade, and periodic timing for each rule.

Until then, `regulatory.thai_sec` must continue to return a visible
`NOT_CONFIGURED` warning rather than a false PASS or an unsupported BLOCK.

## Implementation Boundary

The regulatory program does not authorize fund-optional portfolio development.
The preferred implementation boundary is the `compliance` module: versioned
rule definitions, approved parameters, evidence, effective dates, and
portfolio-scoped bindings. Any new regulatory-profile field or cross-module
contract requires a separate reviewed design; it must not make the existing
portfolio-to-fund association nullable.
