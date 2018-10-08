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
	vpcId = "vpc_id"
)

// EC2SG is an evaluation entity representing an ec2 ebs volume
type EC2SG struct {
	Entity
	ID   string
	Name string
}

// GetID returns the ec2_ebs_vol id
func (e *EC2SG) GetID() string {
	return e.ID
}

// NewEc2EBSVol returns a new ec2 ebs vol entity
func NewEC2SG(sg *ec2.SecurityGroup, region string) *EC2SG {
	entity := &EC2SG{
		Entity: NewEntity(),
	}
	if sg == nil {
		log.Debug("nil sg")
		return entity
	}

	entity.Region = region

	// otherwise populate with more info
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
	entity.AddLabel(vpcId, sg.VpcId)

	return entity
}

// GetConsoleURL will return a url to the AWS console for this volume
func (e *EC2SG) GetConsoleURL() string {
	t := "https://%s.console.aws.amazon.com/ec2/v2/home?region=%s#SecurityGroups:groupId=%s"
	return fmt.Sprintf(t, e.Region, e.Region, e.ID)
}

// Delete deletes
func (e *EC2SG) Delete() error {
	log.Warnf("Would delete ec2_ebs_vol %s", e.ID)
	return nil
}

// EvalEC2SG walks through all ec2 instances
func (c *Client) EvalEC2SG(accounts []*policy.Account, p policy.Policy, regions []string, f func(policy.Violation)) error {
	var errs error
	ctx := context.Background()
	err := c.WalkAccountsAndRegions(accounts, regions, func(client *cziAws.Client, account *policy.Account, region string) {
		var nextToken *string
		for {
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
