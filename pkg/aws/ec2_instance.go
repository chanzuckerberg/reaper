package aws

import log "github.com/sirupsen/logrus"

// ec2_instance specific labels
const (
	ec2InstanceLabelVpcID    = "ec2_instance_vpc_id"
	ec2InstanceLabelPublicIP = "ec2_instance_public_ip"
)

// EC2Instance is an evaluation entity representing an ec2 instance
type EC2Instance struct {
	Entity
	id string
}

// NewEc2Instance returns a new ec2 instance entity
func NewEc2Instance(id string) *EC2Instance {
	return &EC2Instance{
		Entity: NewEntity(),
		id:     id,
	}
}

// Delete deletes
func (e *EC2Instance) Delete() error {
	log.Warnf("Would delete ec2 instance %s", e.id)
	return nil
}

// EC2InstanceClient is an ec2 instance client
type EC2InstanceClient struct {
	EC2Client
}

// func NewEC2InstanceClient(
