package notifier

import (
	"strings"

	"github.com/chanzuckerberg/go-misc/slack"
	"github.com/chanzuckerberg/reaper/pkg/policy"
	slackClient "github.com/nlopes/slack"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Notifier handles sending of notifications
type Notifier struct {
	slack *slack.Client
}

// New will construct a new Notifier and return i
func New(slackToken string) *Notifier {
	api := slack.New(slackToken, log.New())
	return &Notifier{
		slack: api,
	}
}

func parseRecipient(s string) (string, string, error) {
	a := strings.Split(s, "/")
	if len(a) == 2 {
		return a[0], a[1], nil
	}
	return "", "", errors.Errorf("could not parse notification recipient %s", s)
}

func (n *Notifier) lookupUserByUsername(username string) (*slackClient.User, error) {
	users, err := n.slack.Slack.GetUsers()
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		if u.Name == username {
			return &u, nil
		}

	}
	return nil, nil
}

// Send will transmit all violations for the given violation
func (n *Notifier) Send(v policy.Violation) error {
	for _, notif := range v.Policy.Notifications {
		method, rec, _ := parseRecipient(notif.Recipient)
		msg, err := notif.GetMessage(v.Subject, v.Policy)
		if err != nil {
			return err
		}
		if method == "slack" {
			switch string(rec[0]) {
			case "@":
				user, err := n.lookupUserByUsername(rec[1:])
				if err != nil {
					return err
				}
				n.slack.SendMessageToUser(user.ID, msg)
			case "#":
			default:
				return errors.Errorf("don't know how to send to %s", notif.Recipient)
			}
		}
	}
	return nil
}
