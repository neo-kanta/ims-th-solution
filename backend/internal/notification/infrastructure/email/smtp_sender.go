// Package email provides SMTP-based email sending for the notification module.
package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/domain"
	"github.com/neo-kanta/ims-th-solution/backend/platform/config"
)

// SMTPSender implements domain.EmailSender using Go's standard net/smtp package.
// It supports three TLS modes: "none" (plain), "starttls" (STARTTLS upgrade),
// and "tls" (implicit TLS / SMTPS). Provider choice is purely config-driven.
type SMTPSender struct {
	host        string
	port        int
	username    string
	password    string
	tlsMode     string
	fromAddress string
	fromName    string
	timeout     time.Duration
}

// NewSMTPSender builds a sender from application config.
func NewSMTPSender(cfg *config.AppConfig) *SMTPSender {
	return &SMTPSender{
		host:        cfg.SMTPHost,
		port:        cfg.SMTPPort,
		username:    cfg.SMTPUsername,
		password:    cfg.SMTPPassword,
		tlsMode:     cfg.SMTPTLSMode,
		fromAddress: cfg.SMTPFromAddress,
		fromName:    cfg.SMTPFromName,
		timeout:     cfg.SMTPTimeout,
	}
}

// Send delivers msg through SMTP and returns the server's message ID (if any).
func (s *SMTPSender) Send(_ context.Context, msg domain.EmailMessage) (string, error) {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	raw := s.buildRaw(msg)

	switch s.tlsMode {
	case "tls":
		return s.sendTLS(addr, raw)
	case "starttls":
		return s.sendSTARTTLS(addr, raw)
	default:
		return s.sendPlain(addr, raw)
	}
}

func (s *SMTPSender) auth() smtp.Auth {
	if s.username == "" {
		return nil
	}
	return smtp.PlainAuth("", s.username, s.password, s.host)
}

func (s *SMTPSender) sendPlain(addr string, raw []byte) (string, error) {
	conn, err := net.DialTimeout("tcp", addr, s.timeout)
	if err != nil {
		return "", fmt.Errorf("smtp dial: %w", err)
	}
	c, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return "", fmt.Errorf("smtp new client: %w", err)
	}
	defer c.Close()
	return s.send(c, raw)
}

func (s *SMTPSender) sendSTARTTLS(addr string, raw []byte) (string, error) {
	conn, err := net.DialTimeout("tcp", addr, s.timeout)
	if err != nil {
		return "", fmt.Errorf("smtp dial starttls: %w", err)
	}
	c, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return "", fmt.Errorf("smtp new client starttls: %w", err)
	}
	defer c.Close()
	tlsCfg := &tls.Config{ServerName: s.host}
	if err := c.StartTLS(tlsCfg); err != nil {
		return "", fmt.Errorf("smtp starttls: %w", err)
	}
	return s.send(c, raw)
}

func (s *SMTPSender) sendTLS(addr string, raw []byte) (string, error) {
	tlsCfg := &tls.Config{ServerName: s.host}
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: s.timeout}, "tcp", addr, tlsCfg)
	if err != nil {
		return "", fmt.Errorf("smtp dial tls: %w", err)
	}
	c, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return "", fmt.Errorf("smtp new client tls: %w", err)
	}
	defer c.Close()
	return s.send(c, raw)
}

func (s *SMTPSender) send(c *smtp.Client, raw []byte) (string, error) {
	if a := s.auth(); a != nil {
		if err := c.Auth(a); err != nil {
			return "", fmt.Errorf("smtp auth: %w", err)
		}
	}
	if err := c.Mail(s.fromAddress); err != nil {
		return "", fmt.Errorf("smtp mail from: %w", err)
	}
	// Extract single To address from raw message header.
	var to string
	for _, line := range strings.Split(string(raw), "\r\n") {
		if strings.HasPrefix(line, "To: ") {
			// To header may have encoded name; just grab the angle-bracket address.
			to = extractEmail(strings.TrimPrefix(line, "To: "))
			break
		}
	}
	if to == "" {
		return "", fmt.Errorf("smtp: could not determine To address from message")
	}
	if err := c.Rcpt(to); err != nil {
		return "", fmt.Errorf("smtp rcpt to: %w", err)
	}
	wc, err := c.Data()
	if err != nil {
		return "", fmt.Errorf("smtp data: %w", err)
	}
	if _, err := wc.Write(raw); err != nil {
		return "", fmt.Errorf("smtp write: %w", err)
	}
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("smtp data close: %w", err)
	}
	_ = c.Quit()
	return "", nil // standard smtp doesn't return a message ID
}

func (s *SMTPSender) buildRaw(msg domain.EmailMessage) []byte {
	fromFormatted := mime.QEncoding.Encode("utf-8", s.fromName) + " <" + s.fromAddress + ">"
	toFormatted := formatAddress(msg.ToName, msg.ToEmail)

	var b strings.Builder
	b.WriteString("From: " + fromFormatted + "\r\n")
	b.WriteString("To: " + toFormatted + "\r\n")
	b.WriteString("Subject: " + mime.QEncoding.Encode("utf-8", msg.Subject) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")

	if msg.BodyHTML != "" {
		b.WriteString("Content-Type: multipart/alternative; boundary=\"ims-boundary\"\r\n\r\n")
		b.WriteString("--ims-boundary\r\n")
		b.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
		b.WriteString(msg.BodyText + "\r\n")
		b.WriteString("--ims-boundary\r\n")
		b.WriteString("Content-Type: text/html; charset=utf-8\r\n\r\n")
		b.WriteString(msg.BodyHTML + "\r\n")
		b.WriteString("--ims-boundary--\r\n")
	} else {
		b.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
		b.WriteString(msg.BodyText + "\r\n")
	}
	return []byte(b.String())
}

func formatAddress(name, email string) string {
	if name == "" {
		return email
	}
	return mime.QEncoding.Encode("utf-8", name) + " <" + email + ">"
}

func extractEmail(s string) string {
	s = strings.TrimSpace(s)
	// Already a bare email
	if !strings.Contains(s, "<") {
		return s
	}
	start := strings.LastIndex(s, "<")
	end := strings.LastIndex(s, ">")
	if start < 0 || end <= start {
		return s
	}
	return s[start+1 : end]
}

var _ domain.EmailSender = (*SMTPSender)(nil)
