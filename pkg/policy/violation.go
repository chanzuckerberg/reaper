package policy

import (
	"bytes"
	"time"

	"github.com/alecthomas/template"
	units "github.com/docker/go-units"
	"github.com/pkg/errors"
)

// Violation represents a specific resource's lack of compliance to a given policy.
type Violation struct {
	Policy      Policy
	Subject     Subject
	Expired     bool
	AccountID   int64
	AccountName string
}

// NewViolation creates a new Violation struct
func NewViolation(policy Policy, subject Subject, expired bool, accountID int64, accountName string) Violation {
	return Violation{
		Policy:      policy,
		Subject:     subject,
		Expired:     expired,
		AccountID:   accountID,
		AccountName: accountName,
	}
}

// GetMessage gets the notification message
func (v *Violation) GetMessage(s Subject, p Policy) (string, error) {
	message := s.GetID()
	for _, n := range v.Policy.Notifications {
		createdAt := s.GetCreatedAt()
		maxAge := p.MaxAge
		if createdAt != nil && maxAge != nil {
			data := map[string]string{
				"ID":  s.GetID(),
				"Age": units.HumanDuration(time.Since(*createdAt)),
				"TTL": units.HumanDuration(*maxAge - time.Since(*createdAt)),
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
			message = messageBytes.String()
		}
	}
	return message, nil
}
