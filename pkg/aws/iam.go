package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/chanzuckerberg/reaper/pkg/policy"
	log "github.com/sirupsen/logrus"
)

// IAMUser is an evaluation entity representing an ec2 instance
type IAMUser struct {
	Entity
	ID   string
	Name string
}

// GetID returns the ec2_instance id
func (u *IAMUser) GetID() string {
	return fmt.Sprintf("iam:user:%s", u.ID)
}

// GetOwner will return the username as owner
func (u IAMUser) GetOwner() string {
	return u.ID
}

// NewIAMUser returns a new ec2 instance entity
// I don't like that I have to pass accountId and roleName all the way down here.
func (c *Client) NewIAMUser(user *iam.User, accountID int64, roleName string) *IAMUser {
	t := "true"
	entity := &IAMUser{
		Entity: NewEntity(),
	}
	if user == nil {
		return entity
	}

	if user.UserName != nil {
		entity.ID = *user.UserName
	}

	client := c.Get(accountID, roleName, DefaultRegion)
	_, e := client.IAM.GetAnMFASerial(user.UserName)

	if !(e != nil && e.Error() == "No MFA serial Configured") {
		entity.WithLabel("has_mfa", &t)
	}

	login, e := client.IAM.GetLoginProfile(*user.UserName)

	if login != nil {
		entity.WithLabel("has_password", &t)
	}

	return entity
}

// EvalIAMUser walks through all ec2 instances
func (c *Client) EvalIAMUser(accounts []*Account, p policy.Policy, regions []string) ([]policy.Violation, error) {
	var violations []policy.Violation
	var errs error
	for _, account := range accounts {
		log.Infof("Walking iam users for %s", account.Name)
		region := DefaultRegion
		client := c.Get(account.ID, account.Role, region)
		client.IAM.ListAllUsers(func(user *iam.User) {
			i := c.NewIAMUser(user, account.ID, account.Role)
			if p.Match(i) {
				violation := policy.NewViolation(p, i, false, account.ID, account.Name)
				violations = append(violations, violation)
			}
		})

	}
	return violations, errs
}
