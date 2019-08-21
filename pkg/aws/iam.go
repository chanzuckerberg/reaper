package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/chanzuckerberg/reaper/pkg/policy"
	multierror "github.com/hashicorp/go-multierror"
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
	return u.ID
}

// GetOwner will return the username as owner
func (u *IAMUser) GetOwner() string {
	return u.ID
}

// NewIAMUser returns a new ec2 instance entity
// I don't like that I have to pass accountId and roleName all the way down here.
func (c *Client) NewIAMUser(user *iam.User, accountID int64, roleName string, externalID string) *IAMUser {
	ctx := context.Background()
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

	client := c.Get(accountID, roleName, externalID, DefaultRegion)
	_, e := client.IAM.GetAnMFASerial(ctx, user.UserName)

	if !(e != nil && e.Error() == "No MFA serial Configured") {
		entity.AddLabel("has_mfa", &t)
	}

	login, _ := client.IAM.GetLoginProfile(ctx, *user.UserName)
	// if e != nil {
	// 	return errors.Wrap(e, "error fetching login profile")
	// }

	if login != nil {
		entity.AddLabel("has_password", &t)
	}

	return entity
}

// GetConsoleURL will return a URL for this resource in the AWS console
func (u *IAMUser) GetConsoleURL() string {
	t := "https://console.aws.amazon.com/iam/home?region=us-east-1#/users/%s"
	return fmt.Sprintf(t, u.ID)
}

// EvalIAMUser walks through all ec2 instances
func (c *Client) EvalIAMUser(accounts []*policy.Account, p policy.Policy, regions []string) ([]policy.Violation, error) {
	var violations []policy.Violation
	var errs error
	ctx := context.Background()
	for _, account := range accounts {
		log.Infof("Walking iam users for %s", account.Name)
		region := DefaultRegion
		client := c.Get(account.ID, account.Role, account.ExternalID, region)
		err := client.IAM.ListAllUsers(ctx, func(user *iam.User) {
			i := c.NewIAMUser(user, account.ID, account.Role, account.ExternalID)
			if p.Match(i) {
				violation := policy.NewViolation(p, i, false, account)
				violations = append(violations, violation)
			}
		})
		errs = multierror.Append(errs, err)

	}
	return violations, errs
}
