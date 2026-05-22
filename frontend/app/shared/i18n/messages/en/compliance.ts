export const enComplianceMessages = {
  compliance: {
    // Navigation
    nav: {
      section: "Compliance",
      dashboard: "Compliance dashboard",
      rules: "Rule library",
      preTrade: "Pre-trade simulator",
      postTrade: "Post-trade breaches",
      exceptions: "Pre-trade exceptions",
      audit: "Audit trail",
      permissions: "Permission matrix",
      newRule: "New rule",
    },

    // Shared
    common: {
      loading: "Loading…",
      retry: "Retry",
      backendError: "The backend returned an error.",
      none: "—",
      viewMissingApi: "View Missing API checklist",
    },

    // Dashboard
    dashboard: {
      title: "Compliance",
      description:
        "Investment restriction and guideline overview. The pre-trade simulator stops bad orders before they become trade, settlement, or audit problems.",
      kpi: {
        totalRules: "Total rules",
        activeRules: "Active rules",
        inactiveRules: "Inactive / scheduled",
        recentBreaches: "Recent breaches",
      },
      recentFailures: {
        title: "Recent compliance failures",
        description: "Latest open or recently-resolved breach records.",
        empty: "No compliance failures recorded yet.",
      },
      goToSimulator: "Run pre-trade simulator",
      goToRules: "Open rule library",
    },

    // Empty states
    emptyState: {
      noRulesTitle: "No active compliance rules configured",
      noRulesSubtitle: "There is nothing to evaluate orders against.",
      noRulesIntro:
        "Compliance checks need at least one configured rule instance. Until rules are seeded, the pre-trade simulator cannot return a meaningful verdict.",
      setupChecklist: "Setup checklist",
      step1:
        "Confirm migration 20260417000001_compliance__create_rules_tables has been applied (make migrate-up).",
      step2:
        "Create at least one rule via POST /compliance/rules using the swagger payload. The Rule Builder UI ships in Phase 2.",
      step3:
        "Confirm your user holds IRG_VIEW_RULES and WORKFLOW_EXECUTE permissions.",
      noRulesEvaluatedTitle: "Pre-trade check evaluated 0 rules",
      noRulesEvaluatedSubtitle:
        "The backend ran the check but no rule was applicable to this order — a verdict of PASS here would be misleading.",
    },

    // Rule library
    rules: {
      title: "Rule library",
      description:
        "Configured compliance rule instances. Phase 1 read-only; the builder and lifecycle actions arrive in Phase 2.",
      addRuleUnavailable:
        "Rule builder ships in Phase 2. Create rules through the API for now.",
      table: {
        code: "Rule code",
        name: "Name",
        category: "Category",
        severity: "Severity",
        status: "Status",
        version: "Version",
        effective: "Effective window",
        owner: "Owner",
        updated: "Updated",
        empty: "No compliance rules found.",
      },
      filters: {
        title: "Filters",
        ruleTypeId: "Rule type ID",
        ruleTypeIdHint: "e.g. concentration.single_issuer",
        isActive: "Active flag",
        any: "Any",
        activeOnly: "Active only",
        inactiveOnly: "Inactive only",
        apply: "Apply filters",
        reset: "Reset",
      },
    },

    // Pre-trade simulator
    preTrade: {
      title: "Pre-trade compliance simulator",
      description:
        "Submit a proposed order; the backend runs every applicable rule and returns a verdict before any trade is created. A BLOCK verdict here prevents the workflow from advancing.",
      form: {
        sectionOrder: "Proposed order",
        sectionContext: "Portfolio context",
        portfolio: "Portfolio",
        portfolioPlaceholder: "Select a portfolio…",
        portfolioEmpty:
          "No portfolios visible to your user. Ask an administrator to grant a portfolio scope or paste a UUID below.",
        portfolioLoading: "Loading portfolios…",
        portfolioErrorPrefix: "Could not load portfolios:",
        portfolioIdFallback: "Portfolio ID (UUID — fallback)",
        portfolioId: "Portfolio ID",
        contract: "Contract",
        contractPlaceholder: "Select a contract…",
        contractEmpty:
          "No contracts in your permission scope. Paste a UUID below if you know one.",
        contractIdFallback: "Contract ID (UUID — fallback)",
        contractId: "Contract ID",
        businessDate: "Business date",
        orderId: "Order ID",
        generate: "Generate",
        ticker: "Ticker",
        side: "Side",
        buy: "BUY",
        sell: "SELL",
        quantity: "Quantity",
        price: "Price",
        currency: "Currency",
        currencyAutofill: "Auto-filled from selected portfolio.",
        exchange: "Exchange",
        submit: "Run pre-trade check",
        running: "Running compliance check…",
        invalid: "Some required fields are missing or invalid.",
        invalidUuid: "Must be a UUID (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx).",
        invalidDecimal: "Must be a positive decimal number.",
        required: "Required.",
      },
      errors: {
        workflowExecuteRequired:
          "You can view compliance rules, but you do not have WORKFLOW_EXECUTE permission to run pre-trade checks.",
        unauthorizedRules:
          "Your session has expired. Please sign in again.",
      },
      result: {
        block: "BLOCKED",
        blockSummaryOne: "1 rule blocks this order.",
        blockSummaryMany: "{count} rules block this order.",
        warn: "Warning",
        warnSummaryOne: "1 rule raised a warning.",
        warnSummaryMany: "{count} rules raised warnings.",
        pass: "All checks passed",
        passSummary:
          "{count} rules evaluated, 0 breaches. The order is compliant.",
        durationLabel: "Evaluated in {ms} ms",
        groupIdLabel: "Check group",
        rulesEvaluated: "Rules evaluated",
        currentValue: "Current value",
        afterOrder: "After order",
        ruleLimit: "Rule limit",
        difference: "Difference",
        dataNotProvided:
          "Data not returned by the rule — backend evidence object is empty.",
        evidenceLookupFailed:
          "Could not fetch breach evidence from the check group ({error}).",
        explanationLabel: "Why it failed",
        suggestedCorrection: "Suggested correction",
        actions: "Actions",
        recheck: "Recheck",
        requestException: "Request exception",
        requestExceptionDisabled:
          "Exception API not yet implemented (Missing API #10). The disabled state is intentional.",
        viewRule: "View rule",
      },
    },

    // Missing API checklist
    missing: {
      title: "Missing API checklist",
      description:
        "The frontend surfaces every feature that depends on a backend endpoint that does not exist yet. Disabled CTAs in the UI reference the row number here.",
      thNumber: "#",
      thName: "Name",
      thEndpoint: "Endpoint",
      thPurpose: "Purpose",
      thPermission: "Permission",
      thPriority: "Priority",
      thBlocks: "Blocks",
      thPhase: "Phase",
    },

    // Phase 2 — Post-trade breach inbox
    postTrade: {
      title: "Post-trade breach inbox",
      description:
        "Compliance breaches recorded after trade capture or during periodic replay. Override (with reason) is real; full status transitions need backend API #8.",
      empty: "No breaches recorded.",
      filters: {
        title: "Filters",
        status: "Status",
        any: "Any",
        open: "Open",
        overridden: "Overridden",
        resolved: "Resolved",
        ruleTypeId: "Rule type ID",
        dateFrom: "Date from",
        dateTo: "Date to",
        portfolio: "Portfolio",
        portfolioAny: "Any portfolio",
        portfolioId: "Portfolio ID (UUID fallback)",
        portfolioLoading: "Loading portfolios…",
        contract: "Contract",
        contractAny: "Any contract",
        contractId: "Contract ID (UUID fallback)",
        apply: "Apply",
        reset: "Reset",
      },
      table: {
        rule: "Rule",
        severity: "Severity",
        verdict: "Verdict",
        status: "Status",
        businessDate: "Business date",
        portfolio: "Portfolio",
        contract: "Contract",
        message: "Message",
        actions: "Actions",
        override: "Override…",
        viewGroup: "View check group",
        statusTransitionDisabled:
          "Status transitions require Missing API #8 (PATCH /compliance/breaches/{id}).",
      },
      override: {
        title: "Override breach",
        description:
          "Recording an override is permanent and audited. Compliance officers only (IRG_OVERRIDE_BREACH).",
        reason: "Override reason",
        reasonHelp:
          "Minimum 10 characters. Will be stored on the immutable Override record.",
        approvedBy: "Approver (optional)",
        approvedByHelp:
          "Second-level approver UUID, if your control framework requires four-eyes.",
        cancel: "Cancel",
        submit: "Confirm override",
        submitting: "Submitting…",
        success: "Override recorded.",
        notOverridable:
          "This breach is not OPEN. Only OPEN breaches can be overridden.",
      },
      groupDrawer: {
        title: "Check group",
        records: "Check records ({count})",
        breaches: "Breaches ({count})",
        close: "Close",
        loading: "Loading check group…",
        empty: "Backend returned an empty payload.",
      },
    },

    // Phase 2 — Rule Detail
    detail: {
      backToLibrary: "Back to library",
      tabs: {
        overview: "Overview",
        logic: "Logic",
        scope: "Scope",
        test: "Test",
        approval: "Approval",
        audit: "Audit",
        settings: "Settings",
      },
      header: {
        editUnavailable:
          "Edit ships when Missing API #3 (PATCH /compliance/rules/{id}) lands.",
        disableUnavailable:
          "Disable ships when Missing API #4 (POST /compliance/rules/{id}/disable) lands.",
      },
      rail: {
        status: "Status",
        ruleTypeId: "Rule type ID",
        owner: "Owner",
        version: "Active version",
        category: "Category",
        defaultSeverity: "Default severity",
        timings: "Supported timings",
        scopes: "Supported scopes",
        overridable: "Overridable",
        effectiveWindow: "Effective window",
        createdAt: "Created at",
        updatedAt: "Updated at",
      },
      overview: {
        descriptionTitle: "Description",
        descriptionEmpty: "No description provided.",
        metadataDescription: "Description from rule type metadata",
      },
      logic: {
        title: "Logic",
        description:
          "Active parameter version. The Phase 2 builder ships create-only; editing the active version needs Missing API #3.",
        parametersTitle: "Active parameters",
        parametersUnavailable:
          "Active-version payload is not exposed by the list endpoint (Missing API #1 + #2).",
      },
      scope: {
        title: "Scope",
        scopesLabel: "Scope types this rule binds to",
        bindingsTitle: "Bindings",
        bindingsUnavailable:
          "Per-binding overrides are not yet exposed by a backend endpoint.",
      },
      test: {
        title: "Test this rule",
        description:
          "Reuses the live pre-trade simulator. Build an order against the rule's portfolio/contract and read the verdict.",
        cta: "Open in simulator",
      },
      approval: {
        title: "Approval",
        unavailable:
          "Approval flow needs Missing API #5 (submit) and #6 (approve / reject).",
      },
      audit: {
        title: "Audit trail",
        unavailable:
          "Rule version history needs Missing API #2 (GET /compliance/rules/{id}/versions) and #12 (compliance audit log).",
      },
      settings: {
        title: "Settings",
        unavailable:
          "Editing rule metadata, effective_to, or disabling needs Missing API #3 / #4.",
      },
    },

    // Phase 2 — Rule Builder
    builder: {
      title: "New compliance rule",
      description:
        "POST /compliance/rules creates and (optionally) activates a rule. Submit-for-approval, approve/reject, and disable are NOT yet wired — see Missing API #3–#6.",
      lifecycleNotice:
        "This wizard creates a rule and activates it in one step. Full Draft → Pending → Approved lifecycle ships when the supporting endpoints exist.",
      steps: {
        identity: "Identity",
        scope: "Scope",
        logic: "Logic",
        message: "Result message",
        review: "Review",
        submit: "Submit",
      },
      identity: {
        name: "Rule name",
        description: "Description",
        category: "Category",
        ruleTypeId: "Rule type ID",
        ruleTypeIdHelp:
          "Stable backend SPI identifier. Pick from the catalog or type it directly.",
        catalogUnavailable:
          "Rule-type catalog endpoint missing (Missing API #7). Picker uses a static mirror of the seven self-registered backend rule packages.",
      },
      scope: {
        effectiveFrom: "Effective from",
        effectiveTo: "Effective to (optional)",
        isActive: "Activate immediately",
        isActiveHelp:
          "Uncheck to create the instance in an inactive state. Reactivation will need PATCH (Missing API #3).",
      },
      logic: {
        parametersTitle: "Parameters (JSON)",
        parametersHelp:
          "Raw JSON document validated server-side against the rule type's ParameterSchema. Example shown for the selected rule type.",
        invalidJson: "Parameters must be valid JSON.",
      },
      message: {
        title: "Result message",
        notice:
          "Pass/warning/fail messages are produced by the backend rule evaluator today — there is no per-instance override field. This step is intentionally a preview.",
        changeReason: "Change reason",
        changeReasonHelp:
          "Stored on the version record. Useful for audit history (e.g. 'initial creation').",
      },
      review: {
        title: "Review",
        summary: "Summary",
        payloadTitle: "Request payload",
      },
      submit: {
        title: "Submit",
        notice:
          "Submitting calls POST /compliance/rules. The created rule appears in the library immediately.",
        cta: "Create rule",
        submitting: "Creating…",
        success: "Rule created.",
        viewRule: "View rule",
        backToLibrary: "Back to library",
      },
      lifecycleDisabled: {
        submitForApproval: "Submit for approval (Missing API #5)",
        approve: "Approve (Missing API #6)",
        reject: "Reject (Missing API #6)",
        disable: "Disable (Missing API #4)",
        archive: "Archive (Missing API #4)",
      },
      validation: {
        required: "Required.",
        nameTooShort: "At least 3 characters.",
        invalidDate: "Use YYYY-MM-DD.",
      },
    },

    // Phase 2 — Exception Flow
    exceptions: {
      title: "Pre-trade exception requests",
      description:
        "Requests to override a BLOCK verdict before execution. Distinct from post-trade breach overrides.",
      notice:
        "All exception endpoints are missing — see Missing API #9 (list), #10 (request), #11 (approve / reject). The form below is a UX preview; submit is disabled to avoid faking a backend.",
      form: {
        title: "Request exception",
        failedRule: "Failed rule (rule type ID)",
        orderRef: "Order reference",
        portfolioId: "Portfolio ID",
        contractId: "Contract ID",
        reason: "Exception reason",
        justification: "Business justification",
        expiry: "Expiry date",
        approver: "Required approver",
        ack: "I acknowledge the risk and the audit trail.",
        attachmentNote:
          "Attachment upload is not yet supported by the backend.",
        submitDisabled:
          "Cannot submit — Missing API #10 (POST /compliance/exceptions).",
      },
      timeline: {
        title: "Status timeline",
        notRequested: "Not requested",
        draft: "Draft exception",
        pending: "Pending compliance review",
        approved: "Approved exception",
        rejected: "Rejected exception",
        expired: "Expired exception",
      },
    },

    // Phase 2 — Audit Trail
    audit: {
      title: "Audit trail",
      description:
        "Immutable record of compliance actions. The dedicated /compliance/audit endpoint is missing (#12); this view inspects a single check group by id using the real /compliance/checks/{id} API.",
      lookup: {
        title: "Lookup check group",
        groupIdLabel: "Check group UUID",
        cta: "Load",
        loading: "Loading…",
        empty: "Enter a check_group_id from a pre-trade or breach record above.",
      },
      records: "Check records",
      breaches: "Breach records",
      export: {
        label: "Export CSV",
        disabled:
          "CSV export requires Missing API #14 (GET /compliance/audit/export.csv).",
      },
    },

    // Phase 2 — Permission Matrix
    permissions: {
      title: "Permission matrix",
      description:
        "RBAC codes mapped to compliance actions. Sourced from your live PermissionsResp.functions[].",
      youHave: "You hold this permission.",
      youDontHave: "You don't hold this permission.",
      action: "Action",
      code: "Permission code",
      held: "Held",
      table: {
        viewRules: "View compliance rules",
        createRule: "Create rule instance",
        editBinding: "Edit binding",
        overrideBreach: "Override post-trade breach",
        adminRuleType: "Administer rule types",
        runPreTrade: "Run pre-trade / post-trade check",
        viewExceptions: "View exception requests (planned)",
        requestException: "Request pre-trade exception (planned)",
        approveException: "Approve / reject exception (planned)",
      },
    },

    // Investment decision embed
    decision: {
      title: "Investment decision",
      description:
        "Capture a proposed order, run the pre-trade compliance check, and only proceed when the verdict permits.",
      sectionGuide: "Decision workflow",
      step1: "1. Fill the proposed order details.",
      step2: "2. Run the compliance check.",
      step3:
        "3. If BLOCKED, adjust the order or request an exception. If PASS, the decision is ready for approval.",
      readyForApproval: "Compliance cleared — this decision is ready for approval.",
      blockedFromApproval:
        "Compliance blocked this decision. Resolve the breaches before submitting for approval.",
      warnFromApproval:
        "Compliance returned a warning. Approval can proceed with acknowledgement.",
      submitForApprovalDisabled:
        "Approval submission ships in a follow-up — see Missing API #5.",
      submitForApproval: "Submit for approval",
      ownerLabel: "Owner",
      breakdownLabel: "Order breakdown",
    },

    // Badges
    badges: {
      verdict: {
        PASS: "Pass",
        WARN: "Warn",
        BLOCK: "Block",
      },
      severity: {
        BLOCK: "Blocker",
        WARN: "Warning",
        REQUIRE_APPROVAL: "Requires approval",
        MONITOR: "Monitor only",
      },
      status: {
        ACTIVE: "Active",
        SCHEDULED: "Scheduled",
        EXPIRED: "Expired",
        DISABLED: "Disabled",
      },
    },
  },
} as const;
