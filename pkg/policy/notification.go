package policy

import (
	"bytes"
	"strconv"
	"text/template"
	"time"

	units "github.com/docker/go-units"
	"github.com/pkg/errors"
)

// Notification is a notification
type Notification struct {
	MessageTemplate string
	Recipient       string
}

// GetMessage gets the notification message
func (n *Notification) GetMessage(v Violation) (string, error) {
	createdAt := v.Subject.GetCreatedAt()
	maxAge := v.Policy.MaxAge

	data := map[string]string{
		"ID":          v.Subject.GetID(),
		"AccountName": v.AccountName,
		"AccountID":   strconv.FormatInt(v.AccountID, 10),
	}
	if createdAt != nil && maxAge != nil {
		data["Age"] = units.HumanDuration(time.Since(*createdAt))
		data["TTL"] = units.HumanDuration(*maxAge - time.Since(*createdAt))
	}
	t, err := template.New("message").Parse(n.MessageTemplate)
	if err != nil {
		return "", errors.Wrap(err, "Could not create template")
	}

	messageBytes := bytes.NewBuffer(nil)
	err = t.Execute(messageBytes, data)
	if err != nil {
		return "", errors.Wrapf(err, "Could not template message")
	}

	message := messageBytes.String()

	return message, nil
}

// GetRecipient will figure out who should recieve the message
func (n *Notification) GetRecipient(v Violation) string {
	if n.Recipient == "$owner" {
		return v.Subject.GetOwner()
	}
	return n.Recipient
}
