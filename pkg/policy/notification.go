package policy

import (
	"bytes"
	"text/template"
	"time"

	units "github.com/docker/go-units"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Notification is a notification
type Notification struct {
	Slack           *SlackNotification
	MessageTemplate *string
}

// GetMessage gets the notification message
func (n *Notification) GetMessage(s Subject, p Policy) (string, error) {
	message := s.GetID()
	if n.MessageTemplate != nil {
		createdAt := s.GetCreatedAt()
		maxAge := p.MaxAge
		if createdAt != nil && maxAge != nil {
			data := map[string]string{
				"ID":  s.GetID(),
				"Age": units.HumanDuration(time.Since(*createdAt)),
				"TTL": units.HumanDuration(*maxAge - time.Since(*createdAt)),
			}
			t, err := template.New("message").Parse(*n.MessageTemplate)
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

// Notify notifies
func (n *Notification) Notify(s Subject, p Policy) (errs error) {
	message, err := n.GetMessage(s, p)
	if err != nil {
		return err
	}
	log.Warnf("Notify message:\n%s", message)
	if n.Slack != nil {
		errs = multierror.Append(errs, n.Slack.Notify("TODO"))
	}
	return errs
}

// SlackNotification is a slack notification
type SlackNotification struct {
	Channel string
}

// Notify notifies slack
func (sn *SlackNotification) Notify(message string) error {
	// TODO actually notify
	log.Infof("Would notify slack channel %s", sn.Channel)
	return nil
}
