package policy

import (
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
)

// Violation represents a specific resource's lack of compliance to a given policy.
type Violation struct {
	Policy        *Policy
	Subject       Subject
	Expired       bool
	Notifications []Notification
}

// NewViolation creates a new Violation struct
func NewViolation(policy *Policy, subject Subject, expired bool) *Violation {
	return &Violation{
		Policy:  policy,
		Subject: subject,
		Expired: expired,
	}
}

// Notify runs the notification logic on this violation
func (v *Violation) Notify() error {
	var errs error
	for _, n := range v.Notifications {
		err := n.Notify(v.Subject, v.Policy)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}
func (v *Violation) String() string {
	return fmt.Sprintf("subject: %s", v.Subject.GetID())
}
