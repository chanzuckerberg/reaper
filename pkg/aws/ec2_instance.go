package aws

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/chanzuckerberg/aws-tidy/pkg/policy"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
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
	return fmt.Sprintf("ec2_instance: %s", e.ID)
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

// Delete deletes
func (e *EC2Instance) Delete() error {
	log.Warnf("Would delete ec2 instance %s", e.ID)
	return nil
}

// EC2InstanceClient is an ec2 instance client
type EC2InstanceClient struct {
	EC2Client
}

// NewEC2InstanceClient returns new ec2 instance client
func NewEC2InstanceClient(s *session.Session, regions []string, numWorkers int) *EC2InstanceClient {
	ec2Client := NewEC2Client(s, regions, numWorkers)
	return &EC2InstanceClient{*ec2Client}
}

// Walk walks through all ec2 instances
func (e *EC2InstanceClient) Walk(p *policy.Policy, mode string) (errs error) {
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
func (e *EC2InstanceClient) worker(
	wg *sync.WaitGroup,
	p *policy.Policy,
	region string,
	errChan chan<- error) {
	defer wg.Done()
	client, ok := e.RegionClients[region]
	if !ok {
		errChan <- errors.Errorf("EC2 Instance unrecognized region %s", region)
		return
	}

	input := &ec2.DescribeInstancesInput{}
	err := client.DescribeInstancesPages(input, func(output *ec2.DescribeInstancesOutput, lastPage bool) bool {
		for _, reservation := range output.Reservations {
			if reservation == nil {
				continue
			}
			for _, instance := range reservation.Instances {
				ec2InstanceEntity := NewEc2Instance(instance)
				_, err := p.Eval(ec2InstanceEntity)
				if err != nil {
					errChan <- err
				}
			}
		}
		return true
	})
	if err != nil {
		errChan <- errors.Wrap(err, "Error describing ec2 instances")
	}
	return
}
