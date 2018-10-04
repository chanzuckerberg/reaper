package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/chanzuckerberg/reaper/pkg/policy"
	multierror "github.com/hashicorp/go-multierror"
)

// VPC represents an AWS VPC
type VPC struct {
	Entity
	ID   string
	Name string
}

// GetID returns the id of the VPC
func (v *VPC) GetID() string {
	return v.ID
}

// GetConsoleURL will return a URL for this resource in the AWS console
func (v *VPC) GetConsoleURL() string {
	urlTemplate := "https://%s.console.aws.amazon.com/vpc/home?region=%s#vpcs:filter=%s"
	return fmt.Sprintf(urlTemplate, v.Region, v.Region, v.ID)
}

// GetOwner returns the value of the owner tag, if present.
func (v *VPC) GetOwner() string {
	o, ok := v.GetLabels()["owner"]
	if ok {
		return o
	}
	return ""
}

// NewVpc returns a new vpc entity
func NewVpc(vpc *ec2.Vpc, region string) *VPC {
	entity := &VPC{
		Entity: NewEntity(),
	}
	if vpc == nil {
		return entity
	}

	if vpc.VpcId != nil {
		entity.ID = *vpc.VpcId
	}

	entity.Region = region

	for _, tag := range vpc.Tags {
		if tag == nil {
			continue
		}
		if tag.Key != nil && tag.Value != nil && *tag.Key == "Name" {
			entity.Name = *tag.Value
		}
		entity.AddTag(tag.Key, tag.Value)
	}

	entity.AddBoolLabel("is_default", vpc.IsDefault)

	return entity
}

// EvalVPC will evaluate policy for a vpc
func (c *Client) EvalVPC(accounts []*policy.Account, p policy.Policy, regions []string, f func(policy.Violation)) error {
	var errs error
	ctx := context.Background()

	err := c.WalkAccountsAndRegions(accounts, regions, func(client *cziAws.Client, account *policy.Account, region string) {
		err := client.EC2.GetAllVPCs(ctx, func(vpc *ec2.Vpc) {
			v := NewVpc(vpc, region)
			if p.Match(v) {
				violation := policy.NewViolation(p, v, false, account)
				f(violation)
			}

		})
		errs = multierror.Append(errs, err)
	})
	errs = multierror.Append(errs, err)
	return nil
}
