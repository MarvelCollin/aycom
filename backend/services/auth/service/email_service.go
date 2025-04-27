package service

import (
	"fmt"
	"net/smtp"
	"os"
)

// EmailService interface is defined in auth_service.go

// SMTPEmailService implements EmailService using SMTP
type SMTPEmailService struct {
	host     string
	port     string
	username string
	password string
	from     string
}

// NewSMTPEmailService creates a new SMTP email service
func NewSMTPEmailService() EmailService {
	return &SMTPEmailService{
		host:     os.Getenv("SMTP_HOST"),
		port:     os.Getenv("SMTP_PORT"),
		username: os.Getenv("SMTP_USERNAME"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     os.Getenv("SMTP_FROM"),
	}
}

// SendVerificationEmail sends an email with verification code
func (s *SMTPEmailService) SendVerificationEmail(email, code string) error {
	subject := "Verify Your AYCOM Account"
	body := fmt.Sprintf(`
	<html>
	<body>
		<h1>Email Verification</h1>
		<p>Thank you for registering with AYCOM. Please verify your email address to activate your account.</p>
		<p>Your verification code is: <strong>%s</strong></p>
		<p>This code will expire in 5 minutes.</p>
		<p>If you did not request this verification, please ignore this email.</p>
	</body>
	</html>
	`, code)

	return s.sendEmail(email, subject, body)
}

// SendWelcomeEmail sends a welcome email to newly verified users
func (s *SMTPEmailService) SendWelcomeEmail(email, name string) error {
	subject := "Welcome to AYCOM!"
	body := fmt.Sprintf(`
	<html>
	<body>
		<h1>Welcome to AYCOM!</h1>
		<p>Dear %s,</p>
		<p>Thank you for joining AYCOM. Your account has been successfully verified and activated.</p>
		<p>You can now log in and start using our services.</p>
		<p>If you have any questions or need assistance, please don't hesitate to contact our support team.</p>
	</body>
	</html>
	`, name)

	return s.sendEmail(email, subject, body)
}

// sendEmail is a helper function to send an email
func (s *SMTPEmailService) sendEmail(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	// Set up email headers
	headers := make(map[string]string)
	headers["From"] = s.from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// Construct message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Send email
	return smtp.SendMail(addr, auth, s.from, []string{to}, []byte(message))
}
