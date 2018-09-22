package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/chanzuckerberg/reaper/pkg/policy"
	multierror "github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
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
	ID   string
	Name string
}

// GetID returns the ec2_instance id
func (e *EC2Instance) GetID() string {
	return fmt.Sprintf("%s", e.ID)
}

// NewEc2Instance returns a new ec2 instance entity
func NewEc2Instance(instance *ec2.Instance) *EC2Instance {
	entity := &EC2Instance{
		Entity: NewEntity(),
	}
	if instance == nil {
		return entity
	}
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
		entity.WithTag(tag.Key, tag.Value)
	}
	entity.
		WithLabel(ec2InstanceLabelVpcID, instance.VpcId).
		WithLabel(ec2InstanceLabelPublicIP, instance.PublicIpAddress).
		WithLabel(ec2InstanceLabelPrivateIP, instance.PrivateIpAddress).
		WithCreatedAt(instance.LaunchTime)

	return entity
}

// EvalEc2Instance walks through all ec2 instances
func (c *Client) EvalEc2Instance(accounts []*Account, p policy.Policy, regions []string, f func(policy.Violation)) error {
	var errs error
	for _, account := range accounts {
		log.Infof("Walking ec2_instance for %s", account.Name)
		for _, region := range regions {
			log.Infof("scanning %s", region)
			client := c.Get(account.ID, account.Role, region)
			err := client.EC2.GetAllInstances(func(instance *ec2.Instance) {
				i := NewEc2Instance(instance)
				if p.Match(i) {
					violation := policy.NewViolation(p, i, false, account.ID, account.Name)
					f(violation)
				}
			})
			errs = multierror.Append(errs, err)
		}

	}
	return errs
}
