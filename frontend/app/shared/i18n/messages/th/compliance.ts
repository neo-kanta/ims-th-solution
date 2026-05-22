// Phase 1 ships English copy only. The TH catalogue re-exports the EN tree so
// the typed message catalog keeps a consistent shape. The existing core.ts
// already falls back to EN when a localised string is missing, so this keeps
// runtime behaviour identical to "TH not yet translated".
export { enComplianceMessages as thComplianceMessages } from "../en/compliance";
