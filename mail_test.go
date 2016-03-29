package recycleme

import (
	"fmt"
	"testing"
)

func TestSendMail(t *testing.T) {
	_, err := NewEmailConfig("", "", "", "")
	if err == nil {
		t.Error("no error with wrong email config init")
	}

	_, err = NewEmailConfig("host", "recipient", "username", "password")
	if err == nil {
		t.Error("no error with wrong email host init")
	}

	m, err := NewEmailConfig("host:port", "recipient", "username", "password")
	if err != nil {
		t.Fatal(err)
	}
	ec := m.(*emailConfig)
	subject := "subject"
	body := "body"
	expected := fmt.Sprintf("From: %s\nTo: %s\nSubject: [RECYCLEME] %s\n\n%s", ec.sender, ec.recipient, subject, body)
	if msg := ec.createMessage(subject, body); msg != expected {
		t.Errorf("message %v != %v", msg, expected)
	}
}
