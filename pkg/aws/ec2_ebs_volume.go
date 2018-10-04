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

// ec2_instance specific labels
const (
	ec2EBSVolLabelAz          = "az"
	ec2EBSVolLabelIsEncrypted = "is_encrypted"
	ec2EBSVolLabelSize        = "size"
	ec2EBSVolLabelType        = "type"
	ec2EBSVolLabelState       = "state"
)

// EC2EBSVol is an evaluation entity representing an ec2 ebs volume
type EC2EBSVol struct {
	Entity
	ID   string
	Name string
}

// GetID returns the ec2_ebs_vol id
func (e *EC2EBSVol) GetID() string {
	return e.ID
}

// NewEc2EBSVol returns a new ec2 ebs vol entity
func NewEc2EBSVol(vol *ec2.Volume, region string) *EC2EBSVol {
	entity := &EC2EBSVol{
		Entity: NewEntity(),
	}
	if vol == nil {
		return entity
	}

	entity.Region = region

	// otherwise populate with more info
	if vol.VolumeId != nil {
		entity.ID = *vol.VolumeId
	}

	for _, tag := range vol.Tags {
		if tag == nil {
			continue
		}
		if tag.Key != nil && tag.Value != nil && *tag.Key == "Name" {
			entity.Name = *tag.Value
		}
		entity.AddTag(tag.Key, tag.Value)
	}
	entity.
		AddLabel(ec2EBSVolLabelAz, vol.AvailabilityZone).
		AddBoolLabel(ec2EBSVolLabelIsEncrypted, vol.Encrypted).
		AddInt64Label(ec2EBSVolLabelSize, vol.Size).
		AddLabel(ec2EBSVolLabelState, vol.State).
		AddLabel(ec2EBSVolLabelType, vol.VolumeType).
		AddCreatedAt(vol.CreateTime)

	return entity
}

func (e *EC2EBSVol) GetConsoleURL() string {
	t := "https://%s.console.aws.amazon.com/ec2/v2/home?&region=%s#Volumes:search=%s;sort=state"
	return fmt.Sprintf(t, e.Region, e.Region, e.ID)
}

// Delete deletes
func (e *EC2EBSVol) Delete() error {
	log.Warnf("Would delete ec2_ebs_vol %s", e.ID)
	return nil
}

// EvalEbsVolume walks through all ec2 instances
func (c *Client) EvalEbsVolume(accounts []*policy.Account, p policy.Policy, regions []string, f func(policy.Violation)) error {
	var errs error
	ctx := context.Background()
	err := c.WalkAccountsAndRegions(accounts, regions, func(client *cziAws.Client, account *policy.Account, region string) {
		input := &ec2.DescribeVolumesInput{}

		err := client.EC2.Svc.DescribeVolumesPagesWithContext(ctx, input, func(output *ec2.DescribeVolumesOutput, cont bool) bool {
			for _, vol := range output.Volumes {
				v := NewEc2EBSVol(vol, region)
				if p.Match(v) {
					violation := policy.NewViolation(p, v, false, account)
					f(violation)
				}
			}
			return true
		})
		errs = multierror.Append(errs, err)

	})
	errs = multierror.Append(errs, err)

	return errs
}
