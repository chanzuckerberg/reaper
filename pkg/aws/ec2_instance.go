package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/chanzuckerberg/reaper/pkg/policy"
	multierror "github.com/hashicorp/go-multierror"
)

// ec2_instance specific labels
const (
	ec2InstanceLabelVpcID     = "ec2_instance_vpc_id"
	ec2InstanceLabelPublicIP  = "ec2_instance_public_ip"
	ec2InstanceLabelPrivateIP = "ec2_instance_private_ip"
)

// EC2Instance is an evaluation entity representing an ec2 instance
type EC2Instance struct {
	Entity
}

// GetID returns the ec2_instance id
func (e *EC2Instance) GetID() string {
	return e.ID
}

// GetConsoleURL will return a URL for this resource in the AWS console
func (e *EC2Instance) GetConsoleURL() string {
	t := "https://%s.console.aws.amazon.com/ec2/v2/home?&region=%s#Instances:search=%s;sort=desc:instanceState"
	return fmt.Sprintf(t, e.Region, e.Region, e.ID)
}

// NewEc2Instance returns a new ec2 instance entity
func NewEc2Instance(instance *ec2.Instance, region string) *EC2Instance {
	entity := &EC2Instance{
		Entity: NewEntity(),
	}
	if instance == nil {
		return entity
	}

	entity.Region = region
	// otherwise populate with more info
	if instance.InstanceId != nil {
		entity.ID = *instance.InstanceId
	}

	for _, tag := range instance.Tags {
		if tag == nil {
			continue
		}
		if tag.Key != nil && tag.Value != nil && *tag.Key == "Name" {
			entity.Name = *tag.Value
		}
		entity.AddTag(tag.Key, tag.Value)
	}
	entity.
		AddLabel(ec2InstanceLabelVpcID, instance.VpcId).
		AddLabel(ec2InstanceLabelPublicIP, instance.PublicIpAddress).
		AddLabel(ec2InstanceLabelPrivateIP, instance.PrivateIpAddress).
		AddCreatedAt(instance.LaunchTime)

	return entity
}

// EvalEc2Instance walks through all ec2 instances
func (c *Client) EvalEc2Instance(accounts []*policy.Account, p policy.Policy, regions []string, f func(policy.Violation)) error {
	var errs error
	ctx := context.Background()
	err := c.WalkAccountsAndRegions(accounts, regions, func(client *cziAws.Client, account *policy.Account, region string) {
		err := client.EC2.GetAllInstances(ctx, func(instance *ec2.Instance) {
			i := NewEc2Instance(instance, region)
			if p.Match(i) {
				violation := policy.NewViolation(p, i, false, account)
				f(violation)
			}
		})
		errs = multierror.Append(errs, err)
	})
	errs = multierror.Append(errs, err)

	return errs
}
