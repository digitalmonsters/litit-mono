package mail

import (
	"fmt"
	"net/smtp"
	"strings"
)

func NewEmailService(smtpHost, smtpPort, username, password, sendersAddr, senderName string) IEmailService {
	return &EmailService{
		Host:       smtpHost,
		Port:       smtpPort,
		User:       username,
		Password:   password,
		SenderName: senderName,
		SenderMail: sendersAddr,
	}
}

// SendGenericEmail sends a plain-text email.
func (s *EmailService) SendGenericEmail(to, subject, body string) error {
	from := s.SenderMail
	recipient := to
	msg := []byte("From: " + from + "\r\n" +
		"To: " + recipient + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		body + "\r\n")
	auth := smtp.PlainAuth("", s.User, s.Password, s.Host)

	err := smtp.SendMail(s.Host+":"+s.Port, auth, from, []string{recipient}, msg)
	if err != nil {
		return fmt.Errorf("failed to send generic email to %s: %w", recipient, err)
	}

	return nil
}

// SendGenericHTMLEmail sends an HTML email.
func (s *EmailService) SendGenericHTMLEmail(to, subject, body string) error {
	from := s.SenderMail
	recipient := to

	headers := map[string]string{
		"From":         from,
		"To":           recipient,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=\"UTF-8\"",
	}

	var msgBuilder strings.Builder
	for k, v := range headers {
		msgBuilder.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	// Separate headers from body.
	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString(body)

	msg := []byte(msgBuilder.String())

	auth := smtp.PlainAuth("", s.User, s.Password, s.Host)
	err := smtp.SendMail(
		s.Host+":"+s.Port, // address
		auth,
		from,
		[]string{recipient},
		msg,
	)
	if err != nil {
		return fmt.Errorf("failed to send generic HTML email to %s: %w", recipient, err)
	}

	return nil
}
