package recycleme

import (
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

func sendMail(subject, body string) error {
	host, hostOk := os.LookupEnv("RECYLEME_MAIL_HOST")
	recipient, recipientOk := os.LookupEnv("RECYLEME_MAIL_RECIPIENT")
	username, usernameOk := os.LookupEnv("RECYLEME_MAIL_USERNAME")
	password, passwordOk := os.LookupEnv("RECYLEME_MAIL_PASSWORD")
	if !hostOk || !recipientOk || !usernameOk || !passwordOk {
		return errors.New("no mail environment")
	}
	if !strings.Contains(host, ":") {
		return fmt.Errorf("no port specified for host %v", host)
	}
	auth := smtp.PlainAuth("", username, password, strings.Split(host, ":")[0])
	sender := "admin@howtorecycle.me"
	msg := fmt.Sprintf("From: %s\nTo: %s\nSubject: [RECYCLEME] %s\n\n%s", sender, recipient, subject, body)
	return smtp.SendMail(host, auth, sender, []string{recipient}, []byte(msg))
}
