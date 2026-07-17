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
        "Monitor configured compliance rules, triage open breach records, and move directly into each portfolio's compliance workspace.",
      asOfLabel: "As of {date} (your device's local date)",
      headerActions: {
        refresh: "Refresh",
        openBreaches: "Open breaches",
        newRule: "New rule",
      },
      indicators: {
        groupLabel: "Compliance indicators",
      },
      errors: {
        rulesTitle: "Rule list:",
      },
      breachQueue: {
        title: "Open breach queue",
        description: "Newest open records available to your compliance role, prioritised by severity within this page.",
        errorTitle: "Failed to load open breaches",
        empty: "No open breaches.",
        emptyDescription: "The compliance service returned no open breach records for this role.",
        severitySr: "Severity: {severity}",
        portfolioUnavailable: "Portfolio unavailable",
        review: "Review",
        truncated: "Showing {shown} of {total} open breaches.",
      },
      finder: {
        title: "Find a portfolio",
        description: "Jump to a portfolio's compliance workspace by its code or name.",
        inputLabel: "Portfolio code or name",
        placeholder: "e.g. PF-1024 or Growth Fund",
        noMatches: "No portfolios match this search.",
        permissionUnavailable:
          "Portfolio search is unavailable because this account does not have portfolio-view permission.",
      },
      tabs: {
        label: "Compliance sections",
        overview: "Overview",
        library: "Library",
        approvals: "Approvals",
        breaches: "Breaches",
        exceptions: "Exceptions",
        audit: "Audit",
        settings: "Settings",
      },
      kpi: {
        activeRules: "Active definitions",
        activeRulesSub: "configured and locally in date",
        highRisk: "Block-by-default",
        highRiskSub: "rule type default = BLOCK",
        scheduledRules: "Scheduled definitions",
        scheduledRulesSub: "configured for a future date",
        openBreaches: "Open breaches",
        openBreachesSub: "role-visible records",
      },
      categories: {
        title: "Categories",
        empty: "No categorised rules to summarise yet.",
        items: {
          MANDATE: "Mandate",
          RATIO: "Ratio / exposure",
          RESTRICTION: "Restriction lists",
          REGULATORY: "Regulatory",
          HOUSE: "House rules",
          CLIENT: "Client mandate",
          TEMPORAL: "Temporal",
          BEHAVIORAL: "Behavioural",
          UNKNOWN: "Uncategorised",
        },
      },
      truncatedNotice:
        "Showing the first {shown} of {total} rules — derived counts are computed on this window only.",
      goToRules: "Open rule library",
    },

    // Empty states
    emptyState: {
      noRulesTitle: "No compliance rule definitions configured",
      noRulesSubtitle: "There is no configured rule definition to evaluate yet.",
      noRulesIntro:
        "Compliance checks need at least one configured rule instance before they can return a meaningful verdict.",
      setupChecklist: "Setup checklist",
      step1:
        "Confirm migration 20260417000001_compliance__create_rules_tables has been applied (make migrate-up).",
      step2:
        "Create at least one rule from New rule or through POST /compliance/rules.",
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

    // Rule type catalog — stable technical metadata (rule_type_id, category,
    // parameter keys) lives in lib/ruleTypeCatalog.ts; only the translated
    // presentation copy lives here. Unknown rule types fall back to the raw
    // rule_type_id and the backend message rather than guessed copy.
    catalog: {
      allocation: {
        asset_class_max: {
          label: "Maximum asset-class allocation",
          explanation:
            "Portfolio exposure to the configured asset class after this order would exceed the maximum percentage of NAV.",
          suggestedCorrection:
            "Reduce the order or rebalance other holdings so asset-class exposure stays within the configured ceiling.",
        },
        asset_class_min: {
          label: "Minimum asset-class allocation",
          explanation:
            "Portfolio exposure to the configured asset class is below the minimum percentage of NAV required by the mandate.",
          suggestedCorrection:
            "Increase exposure to the asset class, or confirm the mandate permits a temporary shortfall.",
        },
      },
      amount: {
        minimum_trade: {
          label: "Minimum trade amount",
          explanation:
            "The proposed order's notional value is below the minimum trade amount configured for this mandate.",
          suggestedCorrection:
            "Increase the order quantity or price so the notional value meets the minimum trade amount.",
        },
      },
      cash: {
        availability: {
          label: "Available cash check",
          explanation:
            "Buy order requires more cash than is available in the portfolio for the business date.",
          suggestedCorrection:
            "Reduce order quantity, raise cash (sell other positions), or stage the order for after settlement.",
        },
      },
      concentration: {
        single_issuer: {
          label: "Maximum single-issuer exposure",
          explanation:
            "Combined exposure to the issuer (and any grouped parent entity) exceeds the configured % of NAV.",
          suggestedCorrection:
            "Reduce the order so total issuer exposure stays below the configured cap, or rebalance other positions first.",
        },
      },
      credit: {
        min_rating: {
          label: "Minimum credit rating",
          explanation:
            "Instrument's credit rating is below the minimum permitted for this portfolio or mandate.",
          suggestedCorrection:
            "Choose an instrument that meets the minimum rating, or request a written mandate exception before execution.",
        },
      },
      credit_rating: {
        minimum: {
          label: "Minimum credit rating (stub)",
          explanation:
            "Stub credit-rating rule — instrument rating does not satisfy the configured threshold. This rule type is not yet enforced by the backend.",
          suggestedCorrection:
            "This rule type is not yet enforced. Pick a rating-compliant instrument as a precaution, or request an exception with risk acknowledgement.",
        },
      },
      exposure: {
        max_order_percent_aum: {
          label: "Maximum order size (% of AUM)",
          explanation:
            "The proposed order's trade value exceeds the configured maximum percentage of portfolio AUM for a single order.",
          suggestedCorrection:
            "Reduce the order size, or split it into multiple orders across business dates within the configured limit.",
        },
      },
      quantity: {
        min_trading_unit: {
          label: "Minimum trading unit",
          explanation:
            "Order quantity is below the venue's minimum lot or the configured trading-unit floor.",
          suggestedCorrection:
            "Increase the order quantity to a multiple of the minimum lot size for this market.",
        },
        sell_available: {
          label: "Available-to-sell quantity",
          explanation:
            "Sell order exceeds the position currently available to sell (after pending sells / settlement holds).",
          suggestedCorrection:
            "Lower the sell quantity to the available-to-sell figure, or wait for pending sells to settle.",
        },
      },
      ratio: {
        sector_exposure: {
          label: "Maximum sector exposure",
          explanation:
            "Sector exposure after the order would exceed the configured percentage of NAV.",
          suggestedCorrection:
            "Reduce the order, switch to a different sector, or rebalance prior holdings to free up sector capacity.",
        },
      },
      regulatory: {
        thai_sec: {
          label: "Thai SEC regulatory check (stub)",
          explanation:
            "Stub regulatory check — order violates a configured Thai SEC parameter. This rule type is not yet enforced by the backend and always returns a warning.",
          suggestedCorrection:
            "This rule type is not yet enforced. Review the SEC parameter list with compliance before re-submitting, as a precaution.",
        },
      },
      restriction: {
        blacklist: {
          label: "Restricted security blacklist",
          explanation:
            "Ticker is on the active blacklist (sanctions, banned issuers, internal blocks).",
          suggestedCorrection:
            "Select an unrestricted ticker. Blacklist entries are not overridable from the trading desk.",
        },
        whitelist: {
          label: "Whitelist-only investment",
          explanation:
            "Mandate permits only whitelisted securities, and the proposed ticker is not on the list.",
          suggestedCorrection:
            "Pick a ticker from the mandate whitelist, or request an exception to add it.",
        },
        list_enforcement: {
          label: "Restricted-list enforcement",
          explanation:
            "Combined restricted-list rules (whitelist + blacklist) flagged this order.",
          suggestedCorrection:
            "Check both the whitelist and blacklist; pick a permitted ticker or seek a compliance exception.",
        },
      },
      valuation: {
        min_nav: {
          label: "Minimum portfolio NAV",
          explanation:
            "Portfolio NAV has fallen below the configured floor; further trading is restricted until NAV recovers or the mandate is revised.",
          suggestedCorrection:
            "Escalate to the portfolio manager and compliance before proceeding; this order will not be permitted while NAV remains below the floor.",
        },
      },
    },

    // Pre-trade simulator
    preTrade: {
      title: "Pre-trade compliance simulator",
      description:
        "Submit a proposed order; the backend runs every applicable rule and returns a verdict before any trade is created. A BLOCK verdict here prevents the workflow from advancing.",
      readonlyTitle: "Pre-trade execution requires WORKFLOW_EXECUTE permission.",
      readonlyCopy:
        "You can view active rules and existing results, but running a new simulation is disabled for your role. Request WORKFLOW_EXECUTE from your administrator to enable it.",
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
        permissionRequired: "IRG_OVERRIDE_BREACH permission required.",
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
        updated: "Updated",
        owner: "Owner",
        edit: "Edit",
        disable: "Disable",
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
      noBreaches: "No breaches in this check group.",
      fields: {
        timing: "Timing",
        businessDate: "Business date",
        checkedAt: "Checked at",
        checkedBy: "Checked by",
        order: "Order",
        ticker: "Ticker",
        portfolio: "Portfolio",
        ruleVersion: "Rule version",
        dataHash: "Data hash",
        breach: "Breach",
        created: "Created",
      },
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
      planned: "Planned · backend not wired yet",
      yes: "Yes",
      no: "No",
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
      notConfiguredTitle: "Decision workflow not configured",
      notConfiguredCopy:
        "There is no backend endpoint for the investment-decision lifecycle yet. Submit, review and approve actions cannot be performed from this page.",
      notConfiguredAlternative:
        "To run a pre-trade compliance simulation against a real portfolio, use the Pre-trade Simulator. To post a real trade once the workflow allows it, use the fund's Holdings page.",
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
