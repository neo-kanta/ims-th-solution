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
      notYetAvailable: "Not yet supported by the backend.",
    },

    // Dashboard
    dashboard: {
      title: "Compliance",
      description:
        "Investment restriction and guideline overview. The pre-trade simulator stops bad orders before they become trade, settlement, or audit problems.",
      restrictedHint: "Restricted",
      headerActions: {
        recentChanges: "Recent changes",
        openBreaches: "Open breaches",
        newRule: "New rule",
      },
      tabs: {
        overview: "Overview",
        library: "Library",
        approvals: "Approvals",
        breaches: "Breaches",
        exceptions: "Exceptions",
        audit: "Audit",
        settings: "Settings",
      },
      kpi: {
        activeRules: "Active rules",
        activeRulesSub: "published, in effect",
        draft: "Draft",
        draftSub: "not yet submitted",
        draftHint:
          "Backend has no Draft state today; this counts inactive rules with a future effective_from.",
        pendingReview: "Pending review",
        pendingReviewSub: "awaiting approval",
        pendingReviewHint:
          "Approval flow is not yet supported by the backend.",
        disabled: "Disabled",
        disabledSub: "recently retired",
        expired: "Expired",
        expiredSub: "effective_to lapsed",
        highRisk: "High-risk rules",
        highRiskSub: "severity = BLOCK",
        // Legacy keys still consumed by some demos
        totalRules: "Total rules",
        inactiveRules: "Inactive / scheduled",
        recentBreaches: "Recent breaches",
        trendUnavailable:
          "Week-over-week trend not available — no historical KPI snapshot endpoint yet.",
      },
      recentChanges: {
        title: "Recent rule changes",
        viewAll: "View all →",
        empty: "No rule change history available yet.",
      },
      recentFailures: {
        title: "Recent compliance failures",
        description: "Latest open or recently-resolved breach records.",
        empty: "No compliance failures recorded yet.",
        inboxLink: "Open breaches inbox →",
        reviewCta: "Review",
        preTrade: "pre-trade",
        postTrade: "post-trade",
      },
      search: {
        placeholder: "Search rules: name, code, parameter…",
        filterFund: "Fund",
        filterAssetClass: "Asset class",
        filterRuleType: "Rule type",
        filterStatus: "Status",
        filterSeverity: "Severity",
        filterEffective: "Effective ▼",
        filterOwner: "Owner",
        filterApproval: "Approval ▼",
        backendNote:
          "Backend supports filter by rule_type_id + is_active only. Other facets need extensions to GET /compliance/rules.",
      },
      categories: {
        title: "Categories",
        empty: "No categorised rules to summarise yet.",
      },
      highRiskRules: {
        title: "High-risk rules",
        seeAll: "See all →",
        empty: "No rules with default severity BLOCK.",
      },
      truncatedNotice:
        "Showing the first {shown} of {total} rules — derived counts are computed on this window only.",
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
          "Exception request flow is not yet supported by the backend.",
        viewRule: "View rule",
      },
    },

    // Phase 2 — Post-trade breach inbox
    postTrade: {
      title: "Post-trade breach inbox",
      description:
        "Compliance breaches recorded after trade capture or during periodic replay. Override (with reason) is supported; full status transitions are not yet available.",
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
          "Status transitions are not yet supported by the backend.",
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
          "Edit is not yet supported by the backend.",
        disableUnavailable:
          "Disable is not yet supported by the backend.",
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
          "Active parameter version. The builder is create-only; editing the active version is not yet supported.",
        parametersTitle: "Active parameters",
        parametersUnavailable:
          "Active-version payload is not exposed by the list endpoint.",
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
          "Approval flow is not yet supported by the backend.",
      },
      audit: {
        title: "Audit trail",
        unavailable:
          "Rule version history and a compliance audit feed are not yet supported.",
      },
      settings: {
        title: "Settings",
        unavailable:
          "Editing rule metadata, effective_to, or disabling is not yet supported.",
      },
    },

    // Phase 2 — Rule Builder
    builder: {
      title: "New compliance rule",
      description:
        "POST /compliance/rules creates and (optionally) activates a rule. Submit-for-approval, approve/reject, and disable are not yet supported by the backend.",
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
          "No rule-type catalog endpoint yet — the picker uses a static mirror of the self-registered backend rule packages.",
      },
      scope: {
        effectiveFrom: "Effective from",
        effectiveTo: "Effective to (optional)",
        isActive: "Activate immediately",
        isActiveHelp:
          "Uncheck to create the instance in an inactive state. Reactivation is not yet supported.",
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
        submitForApproval: "Submit for approval (not yet supported)",
        approve: "Approve (not yet supported)",
        reject: "Reject (not yet supported)",
        disable: "Disable (not yet supported)",
        archive: "Archive (not yet supported)",
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
        "Exception endpoints are not yet supported by the backend. The form below is a UX preview; submit is disabled to avoid faking a backend.",
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
          "Cannot submit — exception API is not yet available.",
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
          "CSV export is not yet supported by the backend.",
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
        "Approval submission is not yet supported by the backend.",
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
