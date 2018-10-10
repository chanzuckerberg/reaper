package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/chanzuckerberg/reaper/pkg/policy"
	multierror "github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
)

// IAMAccessKey is an evaluation entity representing an ec2 instance
type IAMAccessKey struct {
	Entity
	ID       string
	UserName string
}

// GetID returns the ec2_instance id
func (u *IAMAccessKey) GetID() string {
	return u.ID
}

// GetOwner will return the username as owner
func (u *IAMAccessKey) GetOwner() string {
	return u.UserName
}

// GetConsoleURL will return a URL for this resource in the AWS console
func (u *IAMAccessKey) GetConsoleURL() string {
	t := "https://console.aws.amazon.com/iam/home?region=us-east-1#/users/%s?section=security_credentials"
	return fmt.Sprintf(t, u.UserName)
}

// NewIAMAccessKey returns a new ec2 instance entity
func (c *Client) NewIAMAccessKey(ctx context.Context, key *iam.AccessKeyMetadata) *IAMAccessKey {
	entity := &IAMAccessKey{
		Entity: NewEntity(),
	}

	entity.ID = *key.AccessKeyId
	entity.UserName = *key.UserName

	entity.AddLabel("status", key.Status)

	if key.CreateDate != nil {
		entity.AddCreatedAt(key.CreateDate)
		age := int64(time.Since(*key.CreateDate).Seconds())
		entity.AddInt64Label("age", &age)
		log.Debugf("user: %s age: %d", *key.UserName, age)
	}

	return entity
}

// EvalIAMAccessKey walks through all IAM users' access keys
func (c *Client) EvalIAMAccessKey(accounts []*policy.Account, p policy.Policy) ([]policy.Violation, error) {
	var violations []policy.Violation
	var errs error
	ctx := context.Background()
	for _, account := range accounts {
		log.Infof("Walking iam access key for %s", account.Name)
		region := DefaultRegion
		client := c.Get(account.ID, account.Role, region)
		err := client.IAM.ListAllUsers(ctx, func(user *iam.User) {
			input := &iam.ListAccessKeysInput{
				UserName: user.UserName,
			}
			output, err := client.IAM.Svc.ListAccessKeysWithContext(ctx, input)
			if err != nil {
				errs = multierror.Append(errs, err)
				return
			}
			for _, keyMetadata := range output.AccessKeyMetadata {
				key := c.NewIAMAccessKey(ctx, keyMetadata)
				if p.Match(key) {
					violation := policy.NewViolation(p, key, false, account)
					violations = append(violations, violation)
				}
			}

		})
		errs = multierror.Append(errs, err)

	}
	return violations, errs
}
