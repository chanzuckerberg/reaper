package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
)

const (
	DefaultRegion = "us-east-1" // TODO find this in the sdk
)

// Client is an AWS client
type Client struct {
}

// WalkFun is a walk function over AWS entities
type WalkFun func(*Entity, error) error

// NewClient returns a new aws client
func NewClient(accounts []*Account, regions []string) (*Client, error) {
	return &Client{}, nil
}

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
