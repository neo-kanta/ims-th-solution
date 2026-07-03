package domain

import "context"

// EmailMessage is the provider-independent send request passed to an EmailSender.
type EmailMessage struct {
	ToEmail  string
	ToName   string
	Subject  string
	BodyText string
	BodyHTML string
}

// EmailSender sends a single email through the configured SMTP provider.
// Implementations must be stateless and safe for concurrent use.
type EmailSender interface {
	Send(ctx context.Context, msg EmailMessage) (providerMessageID string, err error)
}
