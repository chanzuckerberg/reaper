package aws

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/chanzuckerberg/reaper/pkg/policy"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ec2_instance specific labels
const (
	ec2EBSVolLabelAz          = "ec2_eba_vol_az"
	ec2EBSVolLabelIsEncrypted = "ec2_eba_vol_is_encrypted"
	ec2EBSVolLabelSize        = "ec2_eba_vol_size"
	ec2EBSVolLabelType        = "ec2_eba_vol_type"
	ec2EBSVolLabelState       = "ec2_eba_vol_state"
)

// EC2EBSVol is an evaluation entity representing an ec2 ebs volume
type EC2EBSVol struct {
	Entity
	ID   string
	Name string
}

// GetID returns the ec2_ebs_vol id
func (e *EC2EBSVol) GetID() string {
	return fmt.Sprintf("ec2_ebs_vol: %s", e.ID)
}

// NewEc2EBSVol returns a new ec2 ebs vol entity
func NewEc2EBSVol(vol *ec2.Volume) *EC2EBSVol {
	entity := &EC2EBSVol{
		Entity: NewEntity(),
	}
	if vol == nil {
		return entity
	}
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
		entity.WithTag(tag.Key, tag.Value)
	}
	entity.
		WithLabel(ec2EBSVolLabelAz, vol.AvailabilityZone).
		WithBoolLabel(ec2EBSVolLabelIsEncrypted, vol.Encrypted).
		WithInt64Label(ec2EBSVolLabelSize, vol.Size).
		WithLabel(ec2EBSVolLabelState, vol.State).
		WithLabel(ec2EBSVolLabelType, vol.VolumeType).
		WithCreatedAt(vol.CreateTime)

	return entity
}

// Delete deletes
func (e *EC2EBSVol) Delete() error {
	log.Warnf("Would delete ec2_ebs_vol %s", e.ID)
	return nil
}

// EC2EBSVolClient is an ec2 ebs client
type EC2EBSVolClient struct {
	EC2Client
}

// NewEC2EBSVolClient returns new ec2 ebs client
func NewEC2EBSVolClient(s *session.Session, regions []string, numWorkers int) *EC2EBSVolClient {
	ec2Client := NewEC2Client(s, regions, numWorkers)
	return &EC2EBSVolClient{*ec2Client}
}

// Walk walks through all ec2 instances
func (e *EC2EBSVolClient) Walk(p *policy.Policy, mode string) (errs error) {
	log.Infof("Walking ec2_instance")
	errChan := make(chan error)
	wg := &sync.WaitGroup{}
	for region := range e.RegionClients {
		wg.Add(1)
		go func(region string) {
			e.worker(wg, p, region, errChan)
		}(region)
	}

	// Wait till all the work is done
	wg.Wait()
	close(errChan)

	// Accumulate errors
	for err := range errChan {
		errs = multierror.Append(errs, err)
	}
	return
}

// worker does the work
// TODO: probably will need to change concurrency model here at some point
func (e *EC2EBSVolClient) worker(
	wg *sync.WaitGroup,
	p *policy.Policy,
	region string,
	errChan chan<- error) {
	defer wg.Done()
	client, ok := e.RegionClients[region]
	if !ok {
		errChan <- errors.Errorf("EC2 ebs vol unrecognized region %s", region)
		return
	}

	input := &ec2.DescribeVolumesInput{}
	err := client.DescribeVolumesPages(input, func(output *ec2.DescribeVolumesOutput, lastPage bool) bool {
		for _, vol := range output.Volumes {
			// ebsVolEntity := NewEc2EBSVol(vol)
			NewEc2EBSVol(vol)
			// _, err := p.Eval(ebsVolEntity)
			// if err != nil {
			// errChan <- err
			// }
		}
		return true
	})
	if err != nil {
		errChan <- errors.Wrap(err, "Error describing ec2 instances")
	}
	return
}
