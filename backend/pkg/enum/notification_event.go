package enum

// NotificationEventType is the backend-owned enum for notification events.
// Business packages import these constants and pass them to the notification
// service — they must not invent free-form strings.
type NotificationEventType string

const (
	NotificationEventApprovalTaskAssigned   NotificationEventType = "APPROVAL_TASK_ASSIGNED"
	NotificationEventApprovalCompleted      NotificationEventType = "APPROVAL_COMPLETED"
	NotificationEventApprovalRejected       NotificationEventType = "APPROVAL_REJECTED"
	NotificationEventWorkflowStuckDay       NotificationEventType = "WORKFLOW_STUCK_DAY"
	NotificationEventAlertThresholdBreached NotificationEventType = "ALERT_THRESHOLD_BREACHED"
	NotificationEventPermissionGranted      NotificationEventType = "PERMISSION_GRANTED"
	NotificationEventEmailTest              NotificationEventType = "EMAIL_TEST"
)
