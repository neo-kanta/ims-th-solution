# IMS External Documentation Coverage Manifest

Date: 2026-07-17 (Asia/Bangkok)

Program: `IMS-UAT-20260717`

Source root: `C:\Users\kanta\Desktop\TH IMS`

## Coverage outcome

**BLOCKED.** The recursive rescan found 48 files. One required path is a
zero-byte, invalid DOCX and cannot be opened, structurally extracted, or
rendered:

`00_Documentation/reference document/developer setup/install golang-migrate guide.docx`

SHA-256:
`E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855`

The documentation gate therefore cannot pass and no application coding is
permitted. Readable paths are still inspected so the next continuation has the
most complete possible manifest.

Final path disposition:

- 36 fully covered: 24 diagram/image files, 11 workbook/deck/Markdown files,
  and 1 sensitive file reviewed under the required redaction rule.
- 11 valid DOCX packages structurally covered but visual/page QA blocked.
- 1 zero-byte invalid DOCX fully blocked.
- All 48 paths appear in this manifest; the overall gate remains **BLOCKED**.

## Inventory summary

| Type | Count |
| --- | ---: |
| DOCX | 13 |
| XLSX | 9 |
| PPTX | 1 |
| Markdown | 1 |
| PNG | 9 |
| XML | 10 |
| Draw.io | 2 |
| Draw.io backup (`.bkp`) | 2 |
| Draw.io temp (`.dtmp`) | 1 |
| **Total** | **48** |

Two byte-identical duplicate groups were found:

1. SHA-256
   `85A51B3D427F0C2D132EE14C8C04C8DD5AE69D1F57174F15BD6A026E72233CD4`
   - Chinese workflow XML
   - English-folder workflow `.bkp`
2. SHA-256
   `306FB862301EFC215ED4B0EDA13DEA925EE2C56CB8D5855850A3574D447844BD`
   - Chinese stock-investment flow XML
   - English-folder stock-investment `.bkp`

## Sensitive-file handling

| Relative path | Type | Bytes | SHA-256 | Duplicate | Method and scope | Requirements / UAT / conflicts | Status |
| --- | --- | ---: | --- | --- | --- | --- | --- |
| `Repository secret.docx` | DOCX | 16,915 | `DE5D90A693CBB5E55059036D1885EFD9B8AF697F8E040F0C1DB8EACC2BD8C003` | Unique | Primary-manager-only local structural review; no content or values emitted, copied, committed, or transmitted | sensitive file reviewed and redacted. | **COVERED (REDACTED)** |

## Blocked path

| Relative path | Type | Bytes | SHA-256 | Duplicate | Method and scope | Requirements / UAT / conflicts | Status |
| --- | --- | ---: | --- | --- | --- | --- | --- |
| `00_Documentation/reference document/developer setup/install golang-migrate guide.docx` | DOCX | 0 | `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855` | Unique | File-length/hash validation; not a valid OOXML package, so no paragraphs, tables, headers, footers, comments, media, pages, or render exist | Content and UAT implications cannot be established. This single path blocks the documentation and application-coding gates. | **BLOCKED** |

## Readable-path coverage

The following path-level rows are completed from the read-only DOCX,
spreadsheet/presentation/Markdown, and XML/Draw.io/PNG work packages.

### XML, Draw.io, backup, and temp files (15/15 covered)

Notation: `P/L/C/V/E` means pages/layers/cells/vertices/edges.

| Relative path | Type | Bytes | SHA-256 / duplicate | Method and scope | Requirements / UAT / conflicts | Status |
| --- | --- | ---: | --- | --- | --- | --- |
| `00_Documentation/reference diagram/Entity Relation Diagram.drawio` | Draw.io | 1,009 | `59403B605106A51DE7A4DB6145ED62486B0F56D6B17D9D2E755FB98A2BD274E2`; unique | Embedded graph parsed; `1/1/4/2/0`; two schema nodes and no relationships | Schema separation only. UAT database ownership/migration isolation; insufficient as an ERD. | **COVERED** |
| `00_Documentation/reference diagram/IMS_PoC_ERD.xml` | XML | 701,483 | `10E24172D48B08BC369FFB5ABE7D5E88E95CDFBE8BD0DAD972176BD32DA12E5F`; unique | Source is malformed by a bare `&`; structure recovered in memory without editing; `1/1/1744/1681/61`, 12 modules, 45 tables | UAT FK/uniqueness/audit/state integrity. Materially stale: contract-centric operational model, no portfolio model, conflicts with current portfolio-first identity and removal of operational `contract_id`. | **COVERED WITH SOURCE DEFECT** |
| `00_Documentation/reference diagram/ims_scope_overview.drawio.xml` | XML | 9,649 | `E5BE81C77F153122E12FE1240E721EA108417B2CB9071F7FE2F7CBEF9F62B445`; unique | Embedded graph parsed; `1/1/31/15/14` | Market import → research/input → pre-trade IRG → decision → execution → review → valuation/watchlist. Missing approval/confirmation/closing gates; duplicate step 5, missing step 4; fund-order wording is stale. | **COVERED** |
| `00_Documentation/reference diagram/Overview/00_Workflow Management Flowchart_v1.0 Eng Ver.xml` | XML | 37,092 | `929F65670B62703453C7A35FCDFEE4426B34361B7067FB16BE68D45660558DA4`; unique | Embedded graph parsed; `1/1/66/40/24`, 27 labels | Day start, investment lifecycle, two approval stamps, manager approval, transaction/accounting closing, cancellation, post-trade IRG and export. UAT every transition/reversal with audit. Less detailed than the 82-cell source. | **COVERED** |
| `00_Documentation/reference diagram/Overview/IMS design.drawio` | Draw.io | 18,738 | `F2316A3F8EB4E246ADFF614FF259D7DA194329B0E786EB00B67F15C5D2FB9566`; unique | Embedded graph parsed; `1/1/55/34/19` | Conceptual auth/gateway/market/investment/portfolio/approval/notification/ETL/IRG UAT. Conflicts with current modular monolith/Go-interface boundaries; contains naming errors. | **COVERED** |
| `00_Documentation/reference diagram/Overview/IMS_Modular_Monolith_Architecture.drawio.xml` | XML | 27,917 | `7223920971C63AA6ECDF2723DFCDDBCA29762F00BF126A203B2D890FA1AC21CD`; unique | Embedded graph parsed; `1/1/80/64/14`, 58 labels | UAT day gate, compliance, approvals, permissions, audit, scheduling and adapter boundaries. Nuxt 3, `/ready`, Cloudflare and OCI are stale/target claims; one missing module node and three orphan edges. | **COVERED** |
| `00_Documentation/reference diagram/Overview/ims_operation_flow.drawio.xml` | XML | 16,070 | `24C72346BDD767F91CF357964A18E1CEB112FD573671F85974CA274F547CC903`; unique | Embedded graph parsed; `1/1/66/37/27` | UAT day start/cancel, pre-trade IRG, execution/review, day end, transaction import, inventory export, closing/post-trade IRG, OMS/custodian idempotency. Manager approval absent; cancel semantics ambiguous. | **COVERED** |
| `00_Documentation/reference document/Function/IMS reading materials Ch Ver/00_Workflow Management Flowchart_v1.0 Ch Ver.xml` | XML | 48,496 | `85A51B3D427F0C2D132EE14C8C04C8DD5AE69D1F57174F15BD6A026E72233CD4`; exact duplicate group A | Embedded graph parsed; `1/1/82/51/29`, 36 labels | Same topology as English workflow. Material translation conflict: Chinese `成交回報` is trade report/confirmation where English says investment review; Chinese `投資指示` is investment instruction where English says investment research. | **COVERED** |
| `00_Documentation/reference document/Function/IMS reading materials Ch Ver/01_StockInvestmentManagementFlowchart_v1.0 Ch Ver.xml` | XML | 5,303 | `306FB862301EFC215ED4B0EDA13DEA925EE2C56CB8D5855850A3574D447844BD`; exact duplicate group B | Base64/raw-DEFLATE/URI graph decoded; `1/1/80/61/17`, 35 labels | Research variants, approval, decision, IRG loop, execution, trade report/carry-forward, multi-account/single-asset and single-account/multi-asset. UAT intraday cancel/price/quantity permissions. | **COVERED** |
| `00_Documentation/reference document/Function/IMS reading materials Eng Ver/.$00_Workflow Management Flowchart_v1.0 Eng Ver.xml.bkp` | BKP | 48,496 | `85A51B3D427F0C2D132EE14C8C04C8DD5AE69D1F57174F15BD6A026E72233CD4`; exact duplicate group A | Shared extraction with identical Chinese workflow; `1/1/82/51/29` | It is Chinese despite the English-directory name. Language must not be inferred from path. Same UAT as the Chinese workflow. | **COVERED (EXACT DUPLICATE)** |
| `00_Documentation/reference document/Function/IMS reading materials Eng Ver/.$00_Workflow Management Flowchart_v1.0 Eng Ver.xml.dtmp` | DTMP | 47,202 | `A31A67B4222738F87D331E6330A2411D2F519989C30E4BB6CADA1F6D0ADF3338`; unique hash, structural duplicate | Embedded graph parsed; `1/1/82/51/29`; final English XML cells, values, styles, geometry and topology match; only viewport differs | Treat as structural duplicate, not an independent requirement source. | **COVERED** |
| `00_Documentation/reference document/Function/IMS reading materials Eng Ver/.$01_StockInvestmentManagementFlowchart_v1.0 Eng Ver.xml.bkp` | BKP | 5,303 | `306FB862301EFC215ED4B0EDA13DEA925EE2C56CB8D5855850A3574D447844BD`; exact duplicate group B | Shared extraction with identical compressed Chinese stock flow; `1/1/80/61/17` | Misnamed English backup; same stock UAT and language warning. | **COVERED (EXACT DUPLICATE)** |
| `00_Documentation/reference document/Function/IMS reading materials Eng Ver/00_Workflow Management Flowchart_v1.0 Eng Ver.xml` | XML | 47,202 | `CD4475E0365240BA7D67D76027EFDA0A88C43414896FA1C3C0D438327B3E79E2`; unique | Embedded graph parsed; `1/1/82/51/29`, 36 labels | Full English start/cancel, instruction/research, decision, execution, confirmation/review, first/second stamps, manager approval/cancel, transaction/accounting closing/cancel, post-trade IRG, export and fund transfer. Translation conflicts above remain unresolved. | **COVERED** |
| `00_Documentation/reference document/Function/IMS reading materials Eng Ver/01_StockInvestmentManagementFlowchart_v1.0 Eng Ver.xml` | XML | 32,239 | `4BD595F83FD60D7F81E553657BC913C199D3EB622E14AEF6970C226DC732982A`; unique | Embedded graph parsed; `1/1/80/61/17`, 35 labels | Same main lifecycle as Chinese stock flow, but one cell ID, five topology/parent relationships and an English source-less edge differ. Resolve before deriving authoritative acceptance criteria. | **COVERED** |
| `GPT/gg.xml` | XML | 14,092 | `2C25B5513275CA56AF4C6BBBE8F8E3BF88B96DBFBEA1124E44F8503F34C14BD1`; unique | Embedded graph parsed; `1/1/46/30/14` | Financial I/O microservices/event-bus target. UAT market/custodian imports, OMS exchange, PAM handoff, schedules/retries and secure supported channels. Conflicts with current monolith/Nuxt 4; DB/Excel/FTP/email need security/idempotency requirements before implementation. | **COVERED** |

### PNG and screenshot files (9/9 covered)

Every PNG was opened and visually inspected at original resolution. Five also
contained parseable Draw.io `mxfile` metadata.

| Relative path | Type | Bytes | SHA-256 / dimensions | Method and scope | Requirements / UAT / conflicts | Status |
| --- | --- | ---: | --- | --- | --- | --- |
| `00_Documentation/operation picture/single securities buy page.png` | PNG | 441,819 | `077B82B85C8EC81FD2345F6EE934FDFDD24B121E2703AF1EF00CDEF1325C2708`; 1531×651 | Original-resolution visual inspection | It is a bond-buy decision, not generic securities: contract/dates/instrument/exchange/currency/face amount/price/rate/research/settlement plus three approvals. UAT calculations, required fields, identifiers, research link, settlement and stages. | **COVERED** |
| `00_Documentation/operation picture/single securities sell page.png` | PNG | 363,797 | `6D0ABD00B043AA7CE3AC38021238DC51F2FD89B5B2EDA6D0C4FC6F8CA66E6049`; 1500×586 | Original-resolution visual inspection | Stock sell with available shares, lot, previous close, currency/rate, settlement, execution/data/review states and manager/trader. UAT oversell, lot/tick, amount recalc, permissions and transitions. | **COVERED** |
| `00_Documentation/operation picture/trading execution page.png` | PNG | 356,735 | `2DC80B4554B9F53BA908EC5871AE1057872F5A80A79025A86AA7C6302BDA96BA`; 1482×551 | Original-resolution visual inspection | Sell decision with execution-status modal, not standalone entry. UAT partial fills, aggregate quantity/average price, report discrepancy, repeat status and over-confirmation prevention. | **COVERED** |
| `00_Documentation/reference diagram/IMS_PoC_ERD_v2.drawio.png` | PNG | 3,138,817 | `E3C6C3CA90B00B59A44140663344F2A17C4109B4C7E1D0C5922E1B2C6C69E090`; 3411×3188 | Original visual plus embedded graph `1/1/1803/1738/63`; 46 tables, 12 modules | Adds employee/IAM relationships versus 45-table XML. UAT all module FKs, but still contract-centric and stale against portfolio identity. Distinct v2, not a render duplicate. | **COVERED** |
| `00_Documentation/reference diagram/ims_scope_overview.drawio.png` | PNG | 117,419 | `5019B1C66E0C38E1327F43707447753519151FF300311E0B5607D0E65510FE93`; 1171×719 | Original visual plus embedded graph `1/1/31/15/14`, matching companion XML | Same scope UAT; duplicate step 5 and missing approval/closing gates are visible. | **COVERED** |
| `00_Documentation/reference diagram/Overview/00_Workflow Management Flowchart_v1.0 Eng Ver.drawio - Copy.png` | PNG | 148,571 | `700BBF82D91A1A61AE140B5D233EC84C9EBCEA210133E0A4B7A6900D3F9DCB6D`; 1092×691 | Original visual plus embedded graph `1/1/66/40/24`, matching Overview XML | Rich approval/closing/cancel/export chain; use over the stale simplified PNG. | **COVERED** |
| `00_Documentation/reference diagram/Overview/00_Workflow Management Flowchart_v1.0 Eng Ver.drawio.png` | PNG | 55,882 | `89C834D4B7E3CC2BC0F661AC3EAADF96FD672D82C593C72A04CF1238C899097D`; 712×691 | Original visual plus compressed graph `1/1/43/28/13` | Stale simplified variant; marks DB/post-trade IRG out of scope and omits confirmation/posting, second stamp, manager approval, separate closings and exports. Must not define UAT alone. | **COVERED** |
| `00_Documentation/reference document/Function/IMS reading materials Eng Ver/TranstationIndex.png` | PNG | 89,508 | `7E0B69AFE57369823FA546036E74AB1EAF94EC44AEBA8E385EACAA914CC9E378`; 898×631 | Original-resolution visual inspection | AI translation of four filenames only. Confirms inventory, not business behavior; translation is non-authoritative and filename is misspelled. | **COVERED** |
| `00_Documentation/reference document/IMS_Modular_Monolith_Architecture.drawio.png` | PNG | 282,824 | `CDC200F1E20BCD613713D86D0BA068EE70CD5B04A22F80990E48793288F00809`; 1563×877 | Original visual plus embedded graph `1/1/80/65/13`, 59 labels | Similar but not identical to XML; wording/topology differ. Nuxt 3, `/ready`, Cloudflare/OCI are stale/target assumptions. | **COVERED** |

### Visual/diagram synthesis

- Business-day start must precede investment activity; cancellation must
  restore the prior valid state using Thai business dates and attributable
  audit evidence.
- The richest lifecycle is instruction/research → decision → pre-trade IRG →
  approval → execution → trade confirmation/report → posting/carry-forward →
  supervisory review, followed by separate reversible transaction and
  accounting closings plus post-trade IRG/export.
- Execution UAT must cover stock-pool scope, lot/tick, cash, inventory,
  classification, currency/settlement, actual-value compliance recheck,
  partial fills and execution-versus-report discrepancies.
- Batch shapes require per-line permissions, compliance, inventory and audit.
- External market/OMS/custodian/PAM paths require idempotency, retries,
  reconciliation, secure transport and failure isolation; diagrams are not
  evidence those integrations exist.
- Highest-risk conflicts are contract/fund-centric legacy identity, translated
  terminology that changes business meaning, simplified diagrams omitting
  mandatory gates, microservices versus modular-monolith targets, stale Nuxt 3
  and deployment labels, ambiguous cancellation targets, and no authoritative
  Thai SEC thresholds.

### Workbooks, presentation, and Markdown (11/11 covered)

All 41 workbook sheets, including hidden sheets, were structurally inspected
for used ranges, values, formulas/cached results, validation, names, tables,
charts and images. The package contained 7,613 nonempty cells and 355 formulas
with no formula/cell-error values. All 52 rendered workbook pages were visually
inspected. All 29 deck slides, including 10 hidden slides, were structurally
extracted (text, tables, shapes, images and exposed notes), rendered with
native PowerPoint after a TIFF blocked the preferred renderer, and visually
inspected. The Markdown file was read completely (114 lines).

| Relative path | Type | Bytes | SHA-256 / duplicate | Method and scope | Requirements / UAT / conflicts | Status |
| --- | --- | ---: | --- | --- | --- | --- |
| `00_Documentation/D01-Project Plan.xlsx` | XLSX | 13,816 | `AD5C52FEEFE271E9677EAD9BC9B321BB4F25F5C0D44E2536233AC907012D4842`; unique | 1 sheet, 1 rendered page; full used-range/structure/formula/visual inspection | High-level 12-month frontend/backend, permissions, integration, workflow, IRG, UAT and production plan. Planning is not completion evidence; visible copy defects reduce authority. | **COVERED** |
| `00_Documentation/IMS_Cloud_BOQ.xlsx` | XLSX | 47,251 | `E0BCC257274A9757A82B8DBA43208C4A92BC8D3F495EE2F1499005BA0C570C73`; unique | 14 sheets, 1,654 cells, 278 formulas; every used area/render inspected | Multi-cloud comparison selects Cloudflare + OCI at about USD 198.05/month and USD 1,188/six months. UAT hosting/security/HA/SSR/rollback assumptions. Pricing is March 2026, unsourced and not externally verified; 6 vs ~6.5 month ambiguity and date locale ambiguity. | **COVERED** |
| `00_Documentation/IMS_Function_Catalog_Master.xlsx` | XLSX | 23,456 | `D0702646E7A05B628CCD713E7DCC1BA04D2AFC041B1E98F51D3627E2CE0330A9`; unique | 4 sheets; 23 functions all active, 16 active validation rules, permission matrix, 30-row legacy catalog; all rendered | UAT active functions, validation and permission mappings. Conflicts with expanded 28-row catalog copies; canonical ownership unresolved. | **COVERED** |
| `00_Documentation/IMS_Function_Catalog_Master_self.xlsx` | XLSX | 23,796 | `2C5EF6B758755A8DE229210F67C8F7BB6885A6D8167DE35C2E1C842D5043AE85`; unique | 4 sheets; 28 total, 23 active, 5 inactive; all structures/renders inspected | Expanded catalog. Tracker says 25 validations although VR001–VR026 provide 26 and the formula omits the final row; VR026 has an unrelated copied description. | **COVERED** |
| `00_Documentation/reference document/feature/IMS_Chat_AI_Financial_Assistant_Project_Documentation_Summary.md` | Markdown | 8,006 | `D159B5E8773B8BA80D485486343876567DFF138AF24629441E24767A9528E6D2`; unique | All 114 lines read | Read-only permission-aware MCP/API chat; UAT traceable numbers, deterministic tools, provenance, prompt-injection defenses, observability and broader tests. Its existing-code claims were not source-verified here. | **COVERED** |
| `00_Documentation/reference document/Project_Plan.xlsx` | XLSX | 32,093 | `FD82A0CE2CD17D318AB50CE2A10CDA53AAE50233852C9DC22752C6CDD688FDAD`; unique | 5 sheets; 59 items/12 sprints, CI/CD, 17 risks, acceptance tracker; every sheet/render inspected | UAT 401/403, 422 validation, full order lifecycle, stale data, valuation date, alert deduplication, immutable audit, containers, CI scans, health/readiness and rollback. Claims 18 acceptance criteria but lists only AC-01–AC-15. | **COVERED** |
| `00_Documentation/reference document/Systemweb Solution Introduction-IMS- English.pptx` | PPTX | 3,465,583 | `965DE1F2A25372AAA3EF96F0BBE28B26E8411EF1D6824C074359FF4A5B98501C`; unique | 29/29 slides rendered and inspected, 19 visible plus 10 hidden; text/tables/shapes/images/notes inspected | Conceptual workflow, electronic approval, IRG and ETL; hidden slides mostly Chinese counterparts. Marketing/concept material has no current version/acceptance state and is not implementation evidence. | **COVERED** |
| `Book1.xlsx` | XLSX | 9,697 | `1AC789BE0F35C740557758D25D772212A88E26122E441890211E1E399DA1C043`; unique | 1 sheet, 1 rendered page; full inspection | Identity capability checklist. Keycloak, FIPS, passkey and protocol claims require product/source validation; UAT federation/MFA/session boundaries and IMS authorization separation. | **COVERED** |
| `Development Schedule.xlsx` | XLSX | 34,444 | `3F49B271DCBFD0FBA6DE94CC12534B09B259C7226137A6A45A6E73464138C738`; unique | 6 sheets including hidden `Lists`; 59 status formulas and 5 validations; all renders inspected | All 59 manual statuses and all 15 listed acceptance criteria are Pending; blank deadlines leave status formulas without cached results. Schedule is not execution evidence. | **COVERED** |
| `endpoint list.xlsx` | XLSX | 17,285 | `035481EA913B8ACB901AC63C5556D7B51D2E7BE039408E226F1B1C598F1090AF`; unique | 2 sheets; endpoint sheet contains five header cells only; security sheet has 30 recommendation rows; both rendered | No endpoint evidence despite project-plan claims of 17 OpenAPI endpoints. Security architecture is recommendation/target material requiring source and runtime proof. | **COVERED** |
| `IMS_Function_Catalog_Master.xlsx` | XLSX | 23,810 | `47D7992DAAD72004D7E25A42446C43D16E003393558E3E59C27887EBF60A5D7D`; unique | 4 sheets; same 28-total/23-active shape as `_self`; all rendered | Differs at VR019–VR021 module labels (`Approval` vs `Approval Flow (4 Eye Module)`). Canonical catalog/version is unresolved. | **COVERED** |

### Office-document synthesis

- Workflow UAT needs business-date/day-state gates, day-start and manager-
  approval prerequisites, post-approval lockout, reversible transitions and
  attributable audit.
- Investment UAT needs research/analysis → decision → pre-trade IRG →
  execution → confirmation/review plus lot/tick, approved-report, cash,
  oversell, limits and server-side permission checks.
- Approval needs maker-checker separation, groups/teams, priority/fallback,
  minimum stamps, delegation and evidentiary electronic signatures.
- Authorization combines function permission and portfolio/data scope;
  external identity federation/MFA does not replace IMS business authorization.
- Integration scope needs mapping, scheduled ETL, monitoring, retries,
  reconciliation, failure isolation and audit across market/order/custodian/
  accounting boundaries.
- Plan/catalog conflicts include missing AC-16–18, 59 pending schedule items,
  an empty endpoint catalog, competing 23/28-row function catalogs, a formula
  omitting VR026, Nuxt 3/Gin/GORM versus current Nuxt 4/Chi/pgx, GCP versus
  Cloudflare/OCI hosting, legacy fund/contract/account identity versus
  portfolio-first codes, and unverified time-sensitive pricing.

### Valid non-sensitive DOCX packages (11/11 structurally covered; visual blocked)

All 11 valid OOXML packages passed ZIP CRC testing. Every paragraph, table,
header/footer story, comment body, tracked-change element, media part,
relationship and embedded object was structurally extracted to external
scratch data. `P/T/H/F/C/R/M/E` means paragraphs/tables/header parts/footer
parts/comments/tracked-change elements/media/embedded OLE objects.

Page counts and page-level layout assertions are unavailable. The canonical
documents-skill renderer failed because `soffice` is absent. A bounded Word 16
read-only export of byte-identical scratch copies remained responsive but did
not emit the first PDF; the fallback was stopped and only its own Word process
was terminated. Media metadata/dimensions were inspected, but page and media
visuals were not.

| Relative path | Type | Bytes | SHA-256 / duplicate | Method and scope | Requirements / UAT / conflicts | Status |
| --- | --- | ---: | --- | --- | --- | --- |
| `00_Documentation/API Documentation/IMS_Bond_Investment_Decision_API_Documentation.docx` | DOCX | 421,631 | `4A64AF11532E4ECBD319801AD990E485775B9C24C04E6B2FC887AFBDBDEAB56B`; unique | `453/22/0/1/0/0/1/0`; all OOXML stories/tables/media extracted; no page render | Fifteen endpoint groups for contract list, decision CRUD/precheck/submit/cancel, bond/reference/research/options/settlement lookup, approval and batch import. UAT state transitions, enum/validation, data scope, audit and batch failure. Contract-centric terminology conflicts with portfolio-first public identity. | **STRUCTURALLY COVERED / VISUAL BLOCKED** |
| `00_Documentation/API Documentation/IMS_Permission_Approval_API_Documentation.docx` | DOCX | 56,807 | `C13014A3C06ACF9E20C48C82F9C5C00A21AEA00A1712FCC0A9846A8898EC38CF`; unique | `1292/38/0/1/0/0/0/0`; full structural extraction; no page render | Permission change requests/items, approval steps, comments/checks/diffs/labels, merge/apply, effective permissions, validation/errors and critical state-machine tests. Conflicts with legacy direct-configuration documents; owner must decide which changes require approval. | **STRUCTURALLY COVERED / VISUAL BLOCKED** |
| `00_Documentation/reference document/Function/IMS reading materials Ch Ver/00_IMS投資管理系統功能說明文件_流程管理_V1.1.docx` | DOCX | 2,288,199 | `4450AA01A30C45D8BED77BC76C9B9A7D2096B9CDE405D532F869B1B44AAFD8C5`; unique | `646/13/1/2/1/0/17/0`; full structure/comments/media metadata; no page render | Workflow, confirmation/carry-forward, discrepancy handling, supervisory review, reversible transitions, business dates and attributable audit. One unresolved comment; differs materially from EN pair. | **STRUCTURALLY COVERED / VISUAL BLOCKED** |
| `00_Documentation/reference document/Function/IMS reading materials Ch Ver/01_IMS投資管理系統功能說明文件_股票投資管理_V1.1.docx` | DOCX | 14,588,006 | `7D798745E8F71001FC64578E253E828CBFB81461ED8168577717F454293A6454`; unique | `3546/56/1/2/1/158/133/1`; revision tracking active; 132 PNG + 1 EMF and one OLE extracted; no page render | Domestic/foreign/basket research, decisions/import, cancel/reprice/requantity, execution, confirmations/import, carry-forward/reporting and IPO. UAT lot/tick, approved research, scope, oversell, workflow/compliance/approval, batch errors and immutable audit. Active revisions make wording non-final. | **STRUCTURALLY COVERED / VISUAL BLOCKED** |
| `00_Documentation/reference document/Function/IMS reading materials Ch Ver/02_IMS投資管理系統功能說明文件_請假與簽核流程管理_V1.1.docx` | DOCX | 3,556,731 | `0DE12E27E77D532B453881C44F45A35D632242A6AA29F92F0BBA73C335E9D7B9`; unique | `1110/26/1/2/1/1/23/0`; tracking active; full structure/media metadata; no page render | Employee/manager leave, delegates/bulk reassignment, approval groups/teams/members/flows and signatures. UAT quorum, priority fallback, group approval, absent-approver delegation, signature and audit. Active revision/comment unresolved. | **STRUCTURALLY COVERED / VISUAL BLOCKED** |
| `00_Documentation/reference document/Function/IMS reading materials Ch Ver/03_IMS投資管理系統功能說明文件-權限管理V1.0.docx` | DOCX | 5,865,654 | `DB9DE61AA18912EAAFB0B4942E5126D78D812985C753CB1F5F11CEF56B29B0B3`; unique | `904/25/1/2/0/0/31/1`; 30 PNG + 1 EMF and one OLE extracted; no page render | Accounts, groups/mapping, function/data permission, change-log query/config, notification configuration and contract-scoped notification/disable. Legacy direct configuration and contract identity conflict with current approval governance/portfolio scope. | **STRUCTURALLY COVERED / VISUAL BLOCKED** |
| `00_Documentation/reference document/Function/IMS reading materials Eng Ver/00_IMSFunctionDescription_WorkflowManagement_V1.1.docx` | DOCX | 2,314,897 | `0548395F4232FD8F60FE26D9DDA54BC92BD36EB1205931DDADED1875787D1F07`; unique | `788/13/1/2/2/0/17/0`; full structural extraction; no page render | Same workflow domain as ZH but 142 more paragraphs and two comments. Added text and unresolved comments mean it is not equivalent to the ZH source. | **STRUCTURALLY COVERED / VISUAL BLOCKED** |
| `00_Documentation/reference document/Function/IMS reading materials Eng Ver/01_IMSFunctionDescription_StockInvestmentManagementV1.1.docx` | DOCX | 14,710,838 | `02985F61C140486499764A29A4F33A34D095BC00A85F4C5410ABF4A08D80A8B3`; unique | `3844/165/1/2/1/114/133/1`; tracking disabled but revisions present; 132 PNG + 1 EMF and one OLE; no page render | Same broad stock lifecycle/UAT as ZH, but 298 more paragraphs and 109 more tables; English TOC retains Chinese names, so translation is partial and not a clean authority. | **STRUCTURALLY COVERED / VISUAL BLOCKED** |
| `00_Documentation/reference document/Function/IMS reading materials Eng Ver/02_IMSFunctionDescription_Leave&ExpenseApprovalWorkflowManagement_V1.1.docx` | DOCX | 3,592,392 | `76DE80A8C9D41F513764EB05A2DCEE7B9F9B773EFB32A4C75B3F89997F7A66D8`; unique | `1242/98/1/2/1/5/23/0`; tracking disabled but revisions present; no page render | Adds explicit English quorum, priority fallback and group-approval rules; 132 more paragraphs and 72 more tables than ZH. Pair requires owner reconciliation. | **STRUCTURALLY COVERED / VISUAL BLOCKED** |
| `00_Documentation/reference document/Function/IMS reading materials Eng Ver/03_IMSFunctionDescription-PermissionsManagementV1.0.docx` | DOCX | 5,889,659 | `1E2C0CEC604041085BD440B7FD7055EE91F50C5102D29B330C1FDB67F490C6E0`; unique | `972/57/1/2/0/0/31/1`; 30 PNG + 1 EMF and one OLE; no page render | Same permission/notification domain as ZH, with 68 more paragraphs and 32 more tables; English TOC retains Chinese content. Reconcile translation and approval-governance conflict. | **STRUCTURALLY COVERED / VISUAL BLOCKED** |
| `00_Documentation/reference document/IMS_Neo.docx` | DOCX | 623,059 | `C1A3AE8DB7B4E8AC6ADDB8381A5210CD648BB2747DDE7478802FF944591534D1`; unique | `871/23/1/2/0/0/3/0`; full structure/media metadata; no page render | PoC scope, SDLC, modular monolith, tests/observability/security, CI/CD, deployment cost, risks and acceptance criteria. Target/planning evidence only, not proof of implementation. | **STRUCTURALLY COVERED / VISUAL BLOCKED** |

### DOCX synthesis

- The EN/ZH pairs are materially different and cannot share extraction:
  paragraph/table/comment/revision counts differ; English TOCs retain Chinese;
  stock and leave ZH retain active tracked changes; unresolved comments remain.
- Stock UAT must include domestic/foreign/basket/IPO shapes, single-target and
  multi-account orders, batch import, cancel/reprice/requantity, lot/tick,
  research eligibility, scope, cash/oversell, compliance/approval, execution,
  confirmation, carry-forward and immutable audit.
- Approval UAT must include leave delegation, bulk reassignment, team/group
  membership, quorum, priority fallback, group approval, maker-checker,
  signatures and attributable evidence.
- Permission UAT must reconcile direct admin configuration with PR-like
  request/review/merge governance, and enforce function plus portfolio/data
  scope on the backend.
- Legacy documents remain contract/fund-centric; durable manager direction is
  portfolio-code-first while fund association remains required.
- Nothing in these files authorizes Thai SEC thresholds. The product regime
  and source-to-rule matrix remain human-gated.
