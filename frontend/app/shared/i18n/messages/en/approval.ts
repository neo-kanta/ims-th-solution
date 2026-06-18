export const enApprovalMessages = {
  approval: {
    nav: {
      section: "Approval",
      inbox: "Approval inbox",
      processes: "Approval processes",
      groups: "Approval groups",
      teams: "Approval teams",
    },
    inbox: {
      title: "Approval Inbox",
      description: "Pending approvals assigned to you.",
      count: "{total} pending task(s)",
    },
    request: {
      title: "Approval Request",
      description: "Review the request, its timeline and sign off.",
      loading: "Loading request…",
    },
    actions: {
      refresh: "Refresh",
      backToInbox: "Back to inbox",
    },
    errors: {
      noPermission: "You do not have permission to perform this action.",
      notFound: "Approval request not found.",
    },
    config: {
      title: "Approval Configuration",
      description: "Configure approval processes, groups and teams.",
      processes: {
        title: "Approval processes",
        description: "Configure approval processes and stages by type and contract.",
        new: "New process",
        listTitle: "Processes",
        empty: "No approval processes configured yet.",
      },
      groups: {
        title: "Approval groups",
        description: "Manage reusable approver groups and members.",
        new: "New group",
        edit: "Edit group",
        listTitle: "Groups",
        empty: "No approval groups configured yet.",
        membersTitle: "Members — ",
      },
      teams: {
        title: "Approval teams",
        description: "Configure per contract/fund approval teams.",
        new: "New team",
        listTitle: "Teams",
        empty: "No approval teams configured yet.",
      },
    },
  },
} as const;
