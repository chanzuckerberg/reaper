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

	data := map[string]interface{}{
		"ResourceID":   v.Subject.GetID(),
		"ResourceName": v.Subject.GetName(),
		"AccountName":  v.AccountName,
		"AccountID":    strconv.FormatInt(v.AccountID, 10),
		"Resource":     v.Subject,
	}
	if createdAt != nil {
		data["Age"] = units.HumanDuration(time.Since(*createdAt))
	}

	if createdAt != nil && maxAge != nil {
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
