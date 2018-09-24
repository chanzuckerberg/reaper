package policy

import (
	"bytes"
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
func (n *Notification) GetMessage(s Subject, p Policy) (string, error) {
	createdAt := s.GetCreatedAt()
	maxAge := p.MaxAge

	data := map[string]string{
		"ID": s.GetID(),
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
