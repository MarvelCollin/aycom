package service

import (
	"fmt"
	"net/smtp"
)

// EmailService defines methods for sending emails
type EmailService interface {
	SendVerificationEmail(to string, code string) error
}

// SMTPEmailService is the SMTP implementation of EmailService
type SMTPEmailService struct {
	host       string
	port       int
	username   string
	password   string
	from       string
	appBaseURL string
}

// NewSMTPEmailService creates a new SMTP email service
func NewSMTPEmailService(host string, port int, username, password, from, appBaseURL string) EmailService {
	return &SMTPEmailService{
		host:       host,
		port:       port,
		username:   username,
		password:   password,
		from:       from,
		appBaseURL: appBaseURL,
	}
}

// SendVerificationEmail sends an email with a verification code
func (s *SMTPEmailService) SendVerificationEmail(to string, code string) error {
	// Set up authentication information
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	// Create email content
	subject := "Verify Your Email for AYCOM"
	body := fmt.Sprintf(`
	<html>
		<body>
			<h1>Welcome to AYCOM!</h1>
			<p>Thank you for registering. Please use the following verification code to complete your registration:</p>
			<h2 style="background-color: #f0f0f0; padding: 10px; text-align: center;">%s</h2>
			<p>Alternatively, you can click on the following link to verify your email:</p>
			<p><a href="%s/verify-email?code=%s&email=%s">Verify Email</a></p>
			<p>If you did not sign up for AYCOM, please ignore this email.</p>
			<p>Thanks,<br>The AYCOM Team</p>
		</body>
	</html>
	`, code, s.appBaseURL, code, to)

	// Compose the message
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n"
	msg := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\n%s\r\n\r\n%s",
		to, s.from, subject, mime, body)

	// Send the email
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", s.host, s.port),
		auth,
		s.from,
		[]string{to},
		[]byte(msg),
	)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// MockEmailService is a mock implementation of EmailService for testing or development
type MockEmailService struct{}

// NewMockEmailService creates a new mock email service
func NewMockEmailService() EmailService {
	return &MockEmailService{}
}

// SendVerificationEmail in the mock service just logs the email instead of sending it
func (s *MockEmailService) SendVerificationEmail(to string, code string) error {
	fmt.Printf("MOCK EMAIL: To: %s, Code: %s\n", to, code)
	return nil
}
