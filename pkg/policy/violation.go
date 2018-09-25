package policy

// Violation represents a specific resource's lack of compliance to a given policy.
type Violation struct {
	Policy      Policy
	Subject     Subject
	Expired     bool
	AccountID   int64
	AccountName string
	Account     *Account
}

// NewViolation creates a new Violation struct
func NewViolation(policy Policy, subject Subject, expired bool, account *Account) Violation {
	return Violation{
		Policy:      policy,
		Subject:     subject,
		Expired:     expired,
		AccountID:   account.ID,
		AccountName: account.Name,
		Account:     account,
	}
}
