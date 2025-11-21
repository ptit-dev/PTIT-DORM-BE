package utils

import (
	"fmt"
	"net/smtp"
)

// SendMail sends an email using Gmail SMTP
func SendMail(smtpHost, smtpPort, sender, password, recipient, subject, body string) error {
	auth := smtp.PlainAuth("", sender, password, smtpHost)
	msg := []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body))
	return smtp.SendMail(smtpHost+":"+smtpPort, auth, sender, []string{recipient}, msg)
}
