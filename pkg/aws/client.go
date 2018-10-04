package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/chanzuckerberg/reaper/pkg/policy"
)

const (
	// DefaultRegion is the AWS region we use for global resources, like IAM
	DefaultRegion = "us-east-1" // TODO find this in the sdk
)

// Client is an AWS client
type Client struct {
}

// WalkFun is a walk function over AWS entities
type WalkFun func(*Entity, error) error

// NewClient returns a new aws client
func NewClient(accounts []*policy.Account, regions []string) (*Client, error) {
	return &Client{}, nil
}

// Get will return a new account, region and role specific AWS client.
func (c *Client) Get(accountID int64, roleName, region string) *cziAws.Client {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	roleCreds := stscreds.NewCredentials(
		sess,
		roleArn(accountID, roleName), func(p *stscreds.AssumeRoleProvider) {
			p.TokenProvider = stscreds.StdinTokenProvider
		},
	)

	conf := &aws.Config{
		Credentials: roleCreds,
		Region:      aws.String(region),
	}
	return cziAws.New(sess).WithAllServices(conf)
}

func roleArn(accountID int64, roleName string) string {
	return fmt.Sprintf("arn:aws:iam::%d:role/%s", accountID, roleName)
}

// WalkAccountsAndRegions will invoke f for each region in each account supplied.
func (c *Client) WalkAccountsAndRegions(accounts []*policy.Account, regions []string, f func(*cziAws.Client, *policy.Account)) error {
	for _, account := range accounts {
		for _, region := range regions {
			client := c.Get(account.ID, account.Role, region)
			f(client, account)
		}
	}
	return nil
}
