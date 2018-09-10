package policy

import (
	"fmt"
	"strings"
	"time"

	"github.com/blend/go-sdk/selector"
	"github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
)

// Subject is gets evaluated by a policy
type Subject interface {
	GetLabels() map[string]string
	GetTags() map[string]string
	GetCreatedAt() *time.Time
	Delete() error
	GetID() string
}

// Notification is a notification
type Notification struct {
	Slack *SlackNotification
}

// Notify notifies
func (n *Notification) Notify() (errs error) {
	if n.Slack != nil {
		errs = multierror.Append(errs, n.Slack.Notify())
	}
	return errs
}

// SlackNotification is a slack notification
type SlackNotification struct {
	Channel string
}

// Notify notifies slack
func (sn *SlackNotification) Notify() error {
	// TODO actually notify
	log.Infof("Would notify slack channel %s", sn.Channel)
	return nil
}

// Policy is an enforcement policy
type Policy struct {
	// ResourceSelector selects on aws services
	ResourceSelector selector.Selector
	// TagSelector selects on aws object tags
	TagSelector selector.Selector
	// LabelSelector selects on custom generated object labels
	LabelSelector selector.Selector
	// MaxAge how old can this object be and still be selected by this policy
	MaxAge        *time.Duration
	Notifications []Notification
}

// String satisfies Stringer interface
func (p *Policy) String() string {
	res := []string{}
	if p.ResourceSelector != nil {
		res = append(res, fmt.Sprintf("ResourceSelector: %s", p.ResourceSelector.String()))
	}
	if p.TagSelector != nil {
		res = append(res, fmt.Sprintf("TagSelector: %s", p.TagSelector.String()))
	}
	if p.LabelSelector != nil {
		res = append(res, fmt.Sprintf("LabelSelector: %s", p.LabelSelector.String()))
	}
	return strings.Join(res, "\n")
}

// Notify runs the notification logic on this resource
func (p *Policy) Notify(s Subject) error {
	log.Warnf("Notify on %s", s.GetID())
	return nil
}

// Enforce enforces this policy
func (p *Policy) Enforce(s Subject) error {
	if p.Match(s) {
		if p.Expired(s) {
			return s.Delete()
		}
		return p.Notify(s)
	}
	return nil
}

// MatchResource determines if we match an aws resource such as s3 or cloudfront
func (p *Policy) MatchResource(resource map[string]string) bool {
	return p.ResourceSelector.Matches(resource)
}

// Match matches a policy against a resource
func (p *Policy) Match(s Subject) bool {
	labelsMatch := false
	if p.LabelSelector != nil {
		labelsMatch = p.LabelSelector.Matches(s.GetLabels())
	}
	tagsMatch := false
	if p.TagSelector != nil {
		tagsMatch = p.TagSelector.Matches(s.GetTags())
	}
	return labelsMatch && tagsMatch
}

// Expired returns true if a resource is older than maxAge
func (p *Policy) Expired(s Subject) bool {
	createdAt := s.GetCreatedAt()
	if p.MaxAge == nil || createdAt == nil {
		return false
	}
	return time.Since(*createdAt) > *p.MaxAge
}

// New returns a new policy
func New() *Policy {
	return &Policy{}
}

// WithTagSelector adds a tag selector
func (p *Policy) WithTagSelector(query string) (*Policy, error) {
	s, err := selector.Parse(query)
	if err != nil {
		return nil, err
	}
	p.TagSelector = s
	return p, nil
}

// WithLabelSelector adds a label selector
func (p *Policy) WithLabelSelector(query string) (*Policy, error) {
	s, err := selector.Parse(query)
	if err != nil {
		return nil, err
	}
	p.LabelSelector = s
	return p, nil
}
