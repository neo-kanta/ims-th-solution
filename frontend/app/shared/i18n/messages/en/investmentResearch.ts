export const enInvestmentResearchMessages = {
  investmentResearch: {
    // Pages / Headers
    title: "Investment Research",
    newReport: "New research report",
    editReport: "Edit research report",
    researchReport: "Research report",
    descriptionList: "Analyst research reports with recommendations and lifecycle tracking.",
    descriptionNew: "Capture an analyst recommendation in a draft report.",
    loading: "Loading research report…",
    loadingList: "Loading research reports…",
    untitledReport: "Untitled report",
    identification: "Identification",
    reportContent: "Report content",

    // Labels & Table columns / Form Fields
    reportNo: "Report no",
    reportDate: "Report date",
    reportDateFrom: "Report date from",
    reportDateTo: "Report date to",
    effectiveDate: "Effective date",
    instrument: "Instrument",
    instrumentCode: "Instrument code",
    instrumentName: "Instrument name",
    instrumentType: "Instrument type",
    market: "Market",
    currency: "Currency",
    recommendation: "Recommendation",
    reportStatus: "Report status",
    review: "Review",
    reviewStatus: "Review status",
    titleField: "Title",
    ownerUserId: "Owner user ID",
    authorUserId: "Author user ID",
    applicableContractId: "Applicable contract ID",
    applicableContract: "Applicable contract",
    ownerAuthor: "Owner / Author",
    typeMarketCurrency: "Type / Market / Currency",
    companyOverview: "Company overview",
    companyOutlook: "Company outlook",
    financialStatus: "Financial status",
    esgComment: "ESG comment",
    investmentAnalysis: "Investment analysis",
    lifecycleNotes: "Lifecycle notes",
    rejectionReason: "Rejection reason",
    postSubmissionNote: "Post-submission note",
    created: "Created",
    lastUpdated: "Last updated",
    timezone: "Timezone",
    timezoneValue: "Asia/Bangkok (UTC+7)",
    by: "by",
    keepCurrent: "Keep current",

    // Filters
    searchPlaceholder: "Report no, ticker, title",
    search: "Search",
    all: "All",
    selectPlaceholder: "Select…",

    // Form help text and errors
    fieldLimits: {
      maxLength: "Max {max} characters",
      minLength: "Investment analysis must be at least {min} characters",
      uuid: "Must be a UUID",
      uuidOptional: "Must be a UUID (or leave empty)",
      currencyCode: "Must be 3 uppercase letters (e.g. THB)",
    },
    required: "Required",
    characterCounter: "{count} / {max}",
    characterCounterAnalysis: "{count} / {min} characters",
    minimumAnalysisPlaceholder: "Minimum {min} characters",
    uuidOptionalPlaceholder: "UUID (optional)",

    // Actions & Buttons
    actions: {
      apply: "Apply",
      reset: "Reset",
      back: "Back",
      edit: "Edit",
      submit: "Submit",
      cancelSubmission: "Cancel submission",
      delete: "Delete",
      cancel: "Cancel",
      createReport: "Create report",
      saveChanges: "Save changes",
      prev: "Previous",
      next: "Next",
      shortcutHintCreate: "Press ⌘/Ctrl + Enter to create",
      shortcutHintSave: "Press ⌘/Ctrl + Enter to save",
    },

    // Confirm dialogs
    confirm: {
      delete: {
        title: "Delete this research report?",
        description: "Soft-deleting the report removes it from the analyst list and the review queue. The audit trail is retained. This cannot be undone from the UI.",
        confirmLabel: "Delete report",
      },
      submit: {
        title: "Submit report for review?",
        description: "The report will move from Not submitted to Submitted and become visible to the review group. You can cancel the submission until the review is completed.",
        confirmLabel: "Submit",
      },
      cancelSubmit: {
        title: "Cancel submission?",
        description: "The report will return to Not submitted and disappear from the review queue. The reviewer's existing comments are retained.",
        confirmLabel: "Cancel submission",
      },
    },

    // Statuses
    status: {
      draft: "Draft",
      active: "Active",
      expired: "Expired",
      rejected: "Rejected",
    },
    reviewStatusValues: {
      notSubmitted: "Not submitted",
      submitted: "Submitted",
      reviewCompleted: "Review completed",
    },
    recommendationValues: {
      buy: "Buy",
      sell: "Sell",
      hold: "Hold",
    },

    // Errors
    errors: {
      loadList: "Failed to load research reports",
      loadDetail: "Failed to load research report",
      create: "Failed to create research report",
      update: "Failed to update research report",
      delete: "Failed to delete research report",
      submit: "Failed to submit research report",
      cancelSubmit: "Failed to cancel research report submission",
    },

    // Empty state
    empty: {
      title: "No research reports yet",
      description: "Create a draft to start tracking analyst recommendations.",
    },

    // Footer paging
    pager: {
      range: "Page {page} · Showing {count} of {total}",
    },
  },
} as const;
