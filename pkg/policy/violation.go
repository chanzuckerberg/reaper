package policy

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
