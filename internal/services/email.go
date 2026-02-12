package services

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"mime/multipart"
	"net/smtp"
	"net/textproto"

	"edvance-assessment/internal/models"
)

// EmailService handles sending emails via AWS SES SMTP
type EmailService struct {
	host     string
	user     string
	password string
	apiKey   string
	mailFrom string
}

// SMTPConfig holds AWS SES SMTP configuration
type SMTPConfig struct {
	APIKey   string
	Host     string
	User     string
	Password string
	MailFrom string
}

// NewEmailService creates a new email service with SMTP settings
func NewEmailService(cfg SMTPConfig) *EmailService {
	return &EmailService{
		host:     cfg.Host,
		user:     cfg.User,
		password: cfg.Password,
		apiKey:   cfg.APIKey,
		mailFrom: cfg.MailFrom,
	}
}

// GenerateCSV creates a CSV from employee data and returns the bytes
func GenerateCSV(employees []models.EmployeeResponse) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	if err := writer.Write([]string{"Timestamp", "Employee Name", "Annual Salary", "Monthly Income Tax"}); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data
	for _, e := range employees {
		record := []string{
			e.TimeStamp,
			e.EmployeeName,
			fmt.Sprintf("%.2f", e.AnnualSalary),
			fmt.Sprintf("%.2f", e.MonthlyIncomeTax),
		}
		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// SendEmployeeReport sends an email with CSV attachment via AWS SES SMTP
func (s *EmailService) SendEmployeeReport(toEmail string, csvData []byte) error {
	// Build MIME message with attachment
	var msg bytes.Buffer
	mimeWriter := multipart.NewWriter(&msg)
	boundary := mimeWriter.Boundary()

	// MIME headers
	msg.Reset()
	msg.WriteString(fmt.Sprintf("From: %s\r\n", s.mailFrom))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	msg.WriteString("Subject: Employee Monthly Payslip Report\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", boundary))

	// Text body
	textHeader := make(textproto.MIMEHeader)
	textHeader.Set("Content-Type", "text/plain; charset=UTF-8")
	textPart, err := mimeWriter.CreatePart(textHeader)
	if err != nil {
		return fmt.Errorf("failed to create text part: %w", err)
	}
	textPart.Write([]byte("Please find attached the employee monthly payslip report.\r\n"))

	// CSV attachment
	attachHeader := make(textproto.MIMEHeader)
	attachHeader.Set("Content-Type", "text/csv; name=\"employees_payslip.csv\"")
	attachHeader.Set("Content-Disposition", "attachment; filename=\"employees_payslip.csv\"")
	attachHeader.Set("Content-Transfer-Encoding", "base64")
	attachPart, err := mimeWriter.CreatePart(attachHeader)
	if err != nil {
		return fmt.Errorf("failed to create attachment part: %w", err)
	}
	attachPart.Write([]byte(base64.StdEncoding.EncodeToString(csvData)))

	mimeWriter.Close()

	// Send via AWS SES SMTP
	auth := smtp.PlainAuth("", s.user, s.password, s.smtpHostOnly())
	err = smtp.SendMail(s.host, auth, s.mailFrom, []string{toEmail}, msg.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	return nil
}

// smtpHostOnly extracts the hostname without port from the SMTP host address
func (s *EmailService) smtpHostOnly() string {
	for i, c := range s.host {
		if c == ':' {
			return s.host[:i]
		}
	}
	return s.host
}
