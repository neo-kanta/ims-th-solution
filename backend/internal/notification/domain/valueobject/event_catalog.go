// Package valueobject holds immutable notification domain value types.
package valueobject

import "github.com/neo-kanta/ims-th-solution/backend/pkg/enum"

// EventDefinition describes a notification event type with user-facing labels.
type EventDefinition struct {
	Type               enum.NotificationEventType
	Label              string
	Category           string
	Severity           string
	DefaultActionLabel string
	TemplateName       string
}

var eventCatalog = map[enum.NotificationEventType]EventDefinition{
	enum.NotificationEventApprovalTaskAssigned: {
		Type:               enum.NotificationEventApprovalTaskAssigned,
		Label:              "Approval task assigned",
		Category:           "APPROVAL",
		Severity:           "INFO",
		DefaultActionLabel: "Open approval request",
		TemplateName:       "approval_task_assigned",
	},
	enum.NotificationEventApprovalCompleted: {
		Type:               enum.NotificationEventApprovalCompleted,
		Label:              "Approval completed",
		Category:           "APPROVAL",
		Severity:           "INFO",
		DefaultActionLabel: "View approval request",
		TemplateName:       "approval_completed",
	},
	enum.NotificationEventApprovalRejected: {
		Type:               enum.NotificationEventApprovalRejected,
		Label:              "Approval rejected",
		Category:           "APPROVAL",
		Severity:           "WARNING",
		DefaultActionLabel: "View approval request",
		TemplateName:       "approval_rejected",
	},
	enum.NotificationEventWorkflowStuckDay: {
		Type:               enum.NotificationEventWorkflowStuckDay,
		Label:              "Workflow day needs attention",
		Category:           "WORKFLOW",
		Severity:           "WARNING",
		DefaultActionLabel: "View workflow",
		TemplateName:       "workflow_stuck_day",
	},
	enum.NotificationEventAlertThresholdBreached: {
		Type:               enum.NotificationEventAlertThresholdBreached,
		Label:              "Alert threshold breached",
		Category:           "ALERT",
		Severity:           "CRITICAL",
		DefaultActionLabel: "View alert",
		TemplateName:       "alert_threshold_breached",
	},
	enum.NotificationEventPermissionGranted: {
		Type:               enum.NotificationEventPermissionGranted,
		Label:              "Permission granted",
		Category:           "PERMISSION",
		Severity:           "INFO",
		DefaultActionLabel: "View permissions",
		TemplateName:       "permission_granted",
	},
	enum.NotificationEventEmailTest: {
		Type:               enum.NotificationEventEmailTest,
		Label:              "Email test",
		Category:           "SYSTEM",
		Severity:           "INFO",
		DefaultActionLabel: "",
		TemplateName:       "email_test",
	},
}

// LookupEvent returns the event definition for the given type.
// Returns a fallback definition with the raw type when unknown.
func LookupEvent(t enum.NotificationEventType) EventDefinition {
	if def, ok := eventCatalog[t]; ok {
		return def
	}
	return EventDefinition{
		Type:     t,
		Label:    string(t),
		Category: "SYSTEM",
		Severity: "INFO",
	}
}

// LookupEventByString resolves by raw string value (e.g. from DB rows).
func LookupEventByString(s string) EventDefinition {
	return LookupEvent(enum.NotificationEventType(s))
}
