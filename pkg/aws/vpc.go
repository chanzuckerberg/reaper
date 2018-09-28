package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ec2"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/chanzuckerberg/reaper/pkg/policy"
	multierror "github.com/hashicorp/go-multierror"
)

type VPC struct {
	Entity
	ID   string
	Name string
}

func (v *VPC) GetID() string {
	return v.ID
}

func (v *VPC) GetOwner() string {
	o, ok := v.GetLabels()["owner"]
	if ok {
		return o
	}
	return ""
}

// NewVpc returns a new vpc entity
func NewVpc(vpc *ec2.Vpc) *VPC {
	entity := &VPC{
		Entity: NewEntity(),
	}
	if vpc == nil {
		return entity
	}

	if vpc.VpcId != nil {
		entity.ID = *vpc.VpcId
	}

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
	c.WalkAccountsAndRegions(accounts, regions, func(client *cziAws.Client, account *policy.Account) {
		err := client.EC2.GetAllVPCs(ctx, func(vpc *ec2.Vpc) {
			v := NewVpc(vpc)
			if p.Match(v) {
				violation := policy.NewViolation(p, v, false, account)
				f(violation)
			}

		})
		errs = multierror.Append(errs, err)
	})
	return nil
}
