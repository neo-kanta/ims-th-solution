package service

import "fmt"

// EmailTemplateData carries the context fields used when rendering email bodies.
type EmailTemplateData struct {
	RecipientName     string
	BusinessReference string
	BusinessTitle     string
	BusinessType      string
	ActionURL         string
	AppBaseURL        string
	CustomSubject     string
	CustomBody        string
}

// EmailTemplateService renders provider-independent email subjects and bodies.
// For the initial implementation, templates are built inline; future work can
// load them from embedded FS or a database-backed template store.
type EmailTemplateService struct {
	appBaseURL string
}

func NewEmailTemplateService(appBaseURL string) *EmailTemplateService {
	return &EmailTemplateService{appBaseURL: appBaseURL}
}

// Render returns subject, plain-text body, and HTML body for the given event type.
func (s *EmailTemplateService) Render(eventType string, d EmailTemplateData) (subject, bodyText, bodyHTML string) {
	if d.AppBaseURL == "" {
		d.AppBaseURL = s.appBaseURL
	}

	switch eventType {
	case "APPROVAL_TASK_ASSIGNED":
		subject = fmt.Sprintf("Approval request %s needs your action", d.BusinessReference)
		bodyText = fmt.Sprintf(
			"Hello %s,\n\nApproval request %s needs your action.\n\nSubject: %s\n\nPlease review at: %s%s\n\nThis is an automated message from IMS Thailand.",
			d.RecipientName, d.BusinessReference, d.BusinessTitle, d.AppBaseURL, d.ActionURL,
		)
		bodyHTML = fmt.Sprintf(
			"<p>Hello %s,</p><p>Approval request <strong>%s</strong> needs your action.</p><p>Subject: %s</p><p><a href=\"%s%s\">Open approval request</a></p><p><small>This is an automated message from IMS Thailand.</small></p>",
			d.RecipientName, d.BusinessReference, d.BusinessTitle, d.AppBaseURL, d.ActionURL,
		)

	case "APPROVAL_COMPLETED":
		subject = fmt.Sprintf("Approval request %s was approved", d.BusinessReference)
		bodyText = fmt.Sprintf(
			"Hello %s,\n\nApproval request %s has been approved.\n\nSubject: %s\n\nView details at: %s%s\n\nThis is an automated message from IMS Thailand.",
			d.RecipientName, d.BusinessReference, d.BusinessTitle, d.AppBaseURL, d.ActionURL,
		)
		bodyHTML = fmt.Sprintf(
			"<p>Hello %s,</p><p>Approval request <strong>%s</strong> has been approved.</p><p>Subject: %s</p><p><a href=\"%s%s\">View approval request</a></p><p><small>This is an automated message from IMS Thailand.</small></p>",
			d.RecipientName, d.BusinessReference, d.BusinessTitle, d.AppBaseURL, d.ActionURL,
		)

	case "APPROVAL_REJECTED":
		subject = fmt.Sprintf("Approval request %s was rejected", d.BusinessReference)
		bodyText = fmt.Sprintf(
			"Hello %s,\n\nApproval request %s has been rejected.\n\nSubject: %s\n\nView details at: %s%s\n\nThis is an automated message from IMS Thailand.",
			d.RecipientName, d.BusinessReference, d.BusinessTitle, d.AppBaseURL, d.ActionURL,
		)
		bodyHTML = fmt.Sprintf(
			"<p>Hello %s,</p><p>Approval request <strong>%s</strong> has been rejected.</p><p>Subject: %s</p><p><a href=\"%s%s\">View approval request</a></p><p><small>This is an automated message from IMS Thailand.</small></p>",
			d.RecipientName, d.BusinessReference, d.BusinessTitle, d.AppBaseURL, d.ActionURL,
		)

	case "WORKFLOW_STUCK_DAY":
		subject = fmt.Sprintf("Workflow day %s needs attention", d.BusinessReference)
		bodyText = fmt.Sprintf(
			"Hello %s,\n\nWorkflow day %s has not progressed and needs attention.\n\n%s\n\nView details at: %s%s\n\nThis is an automated message from IMS Thailand.",
			d.RecipientName, d.BusinessReference, d.BusinessTitle, d.AppBaseURL, d.ActionURL,
		)
		bodyHTML = fmt.Sprintf(
			"<p>Hello %s,</p><p>Workflow day <strong>%s</strong> has not progressed and needs attention.</p><p>%s</p><p><a href=\"%s%s\">View workflow</a></p><p><small>This is an automated message from IMS Thailand.</small></p>",
			d.RecipientName, d.BusinessReference, d.BusinessTitle, d.AppBaseURL, d.ActionURL,
		)

	case "EMAIL_TEST":
		subject = d.CustomSubject
		if subject == "" {
			subject = "IMS email test"
		}
		bodyText = d.CustomBody
		if bodyText == "" {
			bodyText = fmt.Sprintf("Hello %s,\n\nThis is a test email from IMS Thailand.\n\nIf you received this email, the notification email system is working correctly.", d.RecipientName)
		}
		bodyHTML = fmt.Sprintf(
			"<p>Hello %s,</p><p>%s</p><p><small>This is an automated test message from IMS Thailand.</small></p>",
			d.RecipientName, bodyText,
		)

	default:
		subject = fmt.Sprintf("IMS notification: %s", eventType)
		bodyText = fmt.Sprintf("Hello %s,\n\nYou have a new notification: %s\n\nSubject: %s", d.RecipientName, eventType, d.BusinessTitle)
		bodyHTML = fmt.Sprintf("<p>Hello %s,</p><p>You have a new notification: <strong>%s</strong></p><p>Subject: %s</p>", d.RecipientName, eventType, d.BusinessTitle)
	}
	return subject, bodyText, bodyHTML
}
