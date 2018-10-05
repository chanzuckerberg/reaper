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
	slack       *slack.Client
	ui          ui.UI
	identityMap map[string]string
}

// New will construct a new Notifier and return i
func New(slackToken string, ui ui.UI, iMap map[string]string) *Notifier {
	api := slack.New(slackToken, log.New())
	return &Notifier{
		slack:       api,
		ui:          ui,
		identityMap: iMap,
	}
}

// Send will transmit all violations for the given violation
func (n *Notifier) Send(v policy.Violation) error {
	for _, notif := range v.Policy.Notifications {
		msg, err := notif.GetMessage(v)

		if err != nil {
			return errors.Wrap(err, "could not get message for notification")
		}

		recipient, channel, err := n.Recipient(notif, v)
		if err != nil {
			return err
		}

		if n.ui.Prompt(msg, recipient, "slack") {
			if channel {
				params := slackClient.NewPostMessageParameters()
				params.Markdown = true
				resp, _, e := n.slack.Slack.PostMessage(recipient, msg, params)
				if e != nil {
					log.Errorf("error sending to slack %#v", e)
					return errors.Wrap(e, "error sending message to slack")
				}
				log.Infof("slack PostMessage response: %s, err: %#v", resp, err)
			} else {
				err = n.slack.SendMessageToUserByEmail(recipient, msg, []slackClient.Attachment{})
				if err != nil {
					log.Infof("error sending to slack for %s", recipient)
					return errors.Wrapf(err, "could not send message to %s", recipient)
				}

			}
		}
		// TODO sending to channels and owners
	}
	return nil
}

// Recipient is here because it requires querying slack. Second parameter will be true if the recipient is a channel, otherwise a user/IM.
func (n *Notifier) Recipient(notification policy.Notification, v policy.Violation) (string, bool, error) {
	var email string
	if notification.Recipient == "$owner" {
		owner := v.Subject.GetOwner()
		if owner != "" {
			email = owner
		} else {
			email = v.Account.Owner
		}
	} else {
		email = notification.Recipient
	}

	if c, ok := n.identityMap[email]; ok {
		return c, true, nil
	}
	slackChan, err := n.slack.GetSlackChannelID(email)
	if err == nil {
		log.Infof("slackChan: %#v", slackChan)
		return email, false, nil
	}
	return "", false, nil
}
