package sender

import (
	"fmt"
	"net/smtp"
	"os"
)

// EmailSender 구현체
type EmailSender struct {
	User string
	Pass string
}

func NewEmailSender() *EmailSender {
	return &EmailSender{
		User: os.Getenv("GMAIL_USER"),
		Pass: os.Getenv("GMAIL_PASS"),
	}
}

func (e *EmailSender) Send(recipient string, props map[string]string) error {
	// Properties에서 Email에 필요한 필드 추출
	subject := props["subject"]
	body := props["body"]

	if subject == "" || body == "" {
		return fmt.Errorf("missing subject or body")
	}

	if e.User == "" || e.Pass == "" {
		return fmt.Errorf("GMAIL credentials not set")
	}

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", recipient, subject, body))
	auth := smtp.PlainAuth("", e.User, e.Pass, smtpHost)

	return smtp.SendMail(smtpHost+":"+smtpPort, auth, e.User, []string{recipient}, msg)
}
