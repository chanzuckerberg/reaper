package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/chanzuckerberg/reaper/pkg/policy"
	"github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
)

const (
	vpcID = "vpc_id"
)

// EC2SG is an evaluation entity representing an ec2 security group
type EC2SG struct {
	Entity
	ID   string
	Name string
}

// GetID returns the security group id
func (e *EC2SG) GetID() string {
	return e.ID
}

// NewEC2SG returns a new ec2 security group
func NewEC2SG(sg *ec2.SecurityGroup, region string) *EC2SG {
	entity := &EC2SG{
		Entity: NewEntity(),
	}
	if sg == nil {
		log.Debug("nil sg")
		return entity
	}

	entity.Region = region

	if sg.GroupId != nil {
		entity.ID = *sg.GroupId
	}

	for _, tag := range sg.Tags {
		if tag == nil {
			continue
		}
		if tag.Key != nil && tag.Value != nil && *tag.Key == "Name" {
			entity.Name = *tag.Value
		}
		entity.AddTag(tag.Key, tag.Value)
	}
	entity.AddLabel(vpcID, sg.VpcId)

	return entity
}

// GetConsoleURL will return a url to the AWS console for this security group
func (e *EC2SG) GetConsoleURL() string {
	t := "https://%s.console.aws.amazon.com/ec2/v2/home?region=%s#SecurityGroups:groupId=%s"
	return fmt.Sprintf(t, e.Region, e.Region, e.ID)
}

// Delete deletes
func (e *EC2SG) Delete() error {
	log.Warnf("Would delete security group %s", e.ID)
	return nil
}

// EvalEC2SG walks through all ec2 instances
func (c *Client) EvalEC2SG(accounts []*policy.Account, p policy.Policy, regions []string, f func(policy.Violation)) error {
	var errs error
	ctx := context.Background()
	err := c.WalkAccountsAndRegions(accounts, regions, func(client *cziAws.Client, account *policy.Account, region string) {
		var nextToken *string
		// Limiting to 1000 iteration guarantees that we don't get an infinite loop, even if we have
		// a mistake below. Small tradeoff is that if there are greater than 1000*pagesize security
		// groups we won't scan them all.
		for i := 1; i <= 1000; i++ {
			log.Debugf("nextToken: %#v", nextToken)
			input := &ec2.DescribeSecurityGroupsInput{NextToken: nextToken}

			output, err := client.EC2.Svc.DescribeSecurityGroupsWithContext(ctx, input)

			if err != nil {
				errs = multierror.Append(errs, err)
			} else {
				for _, sg := range output.SecurityGroups {
					s := NewEC2SG(sg, region)
					if p.Match(s) {
						violation := policy.NewViolation(p, s, false, account)
						f(violation)
					}
				}
			}
			if output.NextToken == nil {
				break
			}
			nextToken = output.NextToken
		}
	})
	errs = multierror.Append(errs, err)

	return errs
}
