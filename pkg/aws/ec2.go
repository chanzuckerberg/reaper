package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

// EC2Client is an ec2 client with multi region capabilities
type EC2Client struct {
	Client        ec2iface.EC2API
	RegionClients map[string]ec2iface.EC2API
	Session       *session.Session
	numWorkers    int
}

// NewEC2Client returns a new ec2 client
func NewEC2Client(s *session.Session, regions []string, numWorkers int) *EC2Client {
	ec2Client := &EC2Client{
		Client:        ec2.New(s),
		Session:       s,
		RegionClients: map[string]ec2iface.EC2API{},
		numWorkers:    numWorkers,
	}
	for _, region := range regions {
		ec2Client.RegionClients[region] = ec2.New(s, &aws.Config{Region: aws.String(region)})
	}
	return ec2Client
}
