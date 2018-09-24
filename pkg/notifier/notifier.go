package notifier

import (
	"github.com/chanzuckerberg/go-misc/slack"
	"github.com/chanzuckerberg/reaper/pkg/policy"
	"github.com/chanzuckerberg/reaper/pkg/ui"
	slackClient "github.com/nlopes/slack"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Notifier handles sending of notifications
type Notifier struct {
	slack *slack.Client
	ui    ui.UI
}

// New will construct a new Notifier and return i
func New(slackToken string, ui ui.UI) *Notifier {
	api := slack.New(slackToken, log.New())
	return &Notifier{
		slack: api,
		ui:    ui,
	}
}

// Send will transmit all violations for the given violation
func (n *Notifier) Send(v policy.Violation) error {
	for _, notif := range v.Policy.Notifications {
		msg, err := notif.GetMessage(v.Subject, v.Policy)
		if err != nil {
			return errors.Wrap(err, "could not get message for notification")
		}
		switch detectRecipientType(notif.Recipient) {
		case "email":
			err = n.sendSlackToEmail(msg, notif.Recipient)
			if err != nil {
				return err
			}
		case "owner":
			log.Infof("owner: %#v", v.Subject.GetOwner())
			err = n.sendSlackToEmail(msg, v.Subject.GetOwner())
			if err != nil {
				return err
			}
		}
		// TODO sending to channels
	}
	return nil
}

func detectRecipientType(recipient string) string { // TODO change return type to enum maybe
	if recipient == "$owner" {
		return "owner"
	}
	return "email"
}

func (n *Notifier) sendSlackToEmail(msg, recipient string) error {
	if n.ui.Prompt(msg, recipient, "slack") {
		err := n.slack.SendMessageToUserByEmail(recipient, msg, []slackClient.Attachment{})
		if err != nil {
			return errors.Wrapf(err, "could not send message to %s", recipient)
		}
	}
	return nil
}
