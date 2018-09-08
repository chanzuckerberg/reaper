package policy

import (
	"time"

	"github.com/blend/go-sdk/selector"
)

// Subject is gets evaluated by a policy
type Subject interface {
	GetLabels() map[string]string
	GetTags() map[string]string
	GetCreatedAt() *time.Time
	Delete() error
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
	MaxAge *time.Duration
}

// Notify runs the notification logic on this resource
func (p *Policy) Notify() error {
	return nil
}

// Enforce enforces this policy
func (p *Policy) Enforce(s Subject) error {
	if p.Match(s) {
		if p.Expired(s) {
			return s.Delete()
		}
		return p.Notify()
	}
	return nil
}

// MatchResource determines if we match an aws resource such as s3 or cloudfront
func (p *Policy) MatchResource(resource map[string]string) bool {
	return p.ResourceSelector.Matches(resource)
}

// Match matches a policy against a resource
func (p *Policy) Match(s Subject) bool {
	// labelsMatch := false
	// labels := s.GetLabels()
	// if labels != nil {
	// 	labelsMatch = p.LabelSelector.Matches(labels)
	// }

	// tagsMatch := false
	// TODO: double check nil labels

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

// MustWithTagSelector panics if there is an error
func (p *Policy) MustWithTagSelector(query string) *Policy {
	p, err := p.WithTagSelector(query)
	if err != nil {
		panic(err)
	}
	return p
}

// WithTagSelector panics if there is an error
func (p *Policy) WithTagSelector(query string) (*Policy, error) {
	s, err := selector.Parse(query)
	if err != nil {
		return nil, err
	}
	p.TagSelector = s
	return p, nil
}

// MustWithLabelSelector panics if there is an error
func (p *Policy) MustWithLabelSelector(query string) *Policy {
	p, err := p.WithLabelSelector(query)
	if err != nil {
		panic(err)
	}
	return p
}

// WithLabelSelector panics if there is an error
func (p *Policy) WithLabelSelector(query string) (*Policy, error) {
	s, err := selector.Parse(query)
	if err != nil {
		return nil, err
	}
	p.LabelSelector = s
	return p, nil
}
