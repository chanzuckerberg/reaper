package policy

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/labels"
)

// Subject is gets evaluated by a policy
type Subject interface {
	Delete() error
	GetCreatedAt() *time.Time
	GetID() string
	GetLabels() labels.Set
	GetName() string
	GetOwner() string
	GetTags() labels.Set
	GetConsoleURL() string
	GetRegion() string
}

// Policy is an enforcement policy
type Policy struct {
	Name string
	// ResourceSelector selects on aws services
	ResourceSelector labels.Selector
	// TagSelector selects on aws object tags
	TagSelector labels.Selector
	// LabelSelector selects on custom generated object labels
	LabelSelector labels.Selector
	// MaxAge how old can this object be and still be selected by this policy
	MaxAge        *time.Duration
	Notifications []Notification
}

// String satisfies Stringer interface
func (p *Policy) String() string {
	res := []string{}
	res = append(res, fmt.Sprintf("Name: %s", p.Name))
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

// MatchResource determines if we match an aws resource such as s3 or cloudfront
func (p *Policy) MatchResource(resource labels.Set) bool {
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
	s, err := labels.Parse(query)
	if err != nil {
		return nil, err
	}
	p.TagSelector = s
	return p, nil
}

// AddLabelSelector adds a label selector
func (p *Policy) AddLabelSelector(query string) (*Policy, error) {
	s, err := labels.Parse(query)
	if err != nil {
		return nil, err
	}
	p.LabelSelector = s
	return p, nil
}
