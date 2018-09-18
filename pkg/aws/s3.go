package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/chanzuckerberg/aws-tidy/pkg/policy"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// s3 specific labels
const (
	s3LabelIsPublic TypeEntityLabel = "s3_is_public"
)

// S3Bucket is an evaluation entity representing an s3 bucket
type S3Bucket struct {
	Entity
	name string
}

// NewS3Bucket returns a new s3 bucket entity
func NewS3Bucket(name string) *S3Bucket {
	return &S3Bucket{
		Entity: NewEntity(),
		name:   name,
	}
}

// Delete deletes this bucket
func (s *S3Bucket) Delete() error {
	log.Warnf("Would delete bucket %s", s.name)
	return nil
}

// GetID returns the s3 bucket id
func (s *S3Bucket) GetID() string {
	return fmt.Sprintf("s3:%s", s.name)
}

// S3Client is an s3 client
type S3Client struct {
	Client s3iface.S3API
	// Per region clients
	RegionClients map[string]s3iface.S3API
	Session       *session.Session
	numWorkers    int
}

// NewS3Client returns a new s3 client
func NewS3Client(s *session.Session, regions []string, numWorkers int) *S3Client {
	s3Client := &S3Client{
		Client:        s3.New(s),
		Session:       s,
		RegionClients: map[string]s3iface.S3API{},
		numWorkers:    numWorkers,
	}
	for _, region := range regions {
		s3Client.RegionClients[region] = s3.New(s, &aws.Config{Region: aws.String(region)})
	}
	return s3Client
}

// Eval walks through all s3 buckets
func (s *S3Client) Eval(p *policy.Policy) ([]*policy.Violation, error) {
	log.Infof("Walking s3 buckets")
	var violations []*policy.Violation
	var errs error

	input := &s3.ListBucketsInput{}
	output, err := s.Client.ListBuckets(input)
	if err != nil {
		return nil, errors.Wrap(err, "Could not list buckets")
	}

	for _, bucket := range output.Buckets {
		res, err := s.DescribeBucket(bucket)
		// accumulate errors
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
		if res == nil {
			log.Debugf("Nil bucket - nothing to do")
			continue
		}
		violation, err := p.Eval(res)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
		if violation != nil {
			violations = append(violations, violation)
		}
	}
	return violations, errs
}

// DescribeBucket describes this bucket
func (s *S3Client) DescribeBucket(b *s3.Bucket) (*S3Bucket, error) {
	if b.Name == nil {
		return nil, errors.New("Nil bucket name")
	}
	log.Debugf("Describing bucket %s", *b.Name)
	name := *b.Name
	bucket := NewS3Bucket(name)
	bucket.WithCreatedAt(b.CreationDate)

	locationInput := &s3.GetBucketLocationInput{}
	locationInput.SetBucket(*b.Name)
	location, err := s.Client.GetBucketLocation(locationInput)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get bucket %s location", *b.Name)
	}
	if location.LocationConstraint == nil {
		// why we can't have nice things
		location.LocationConstraint = aws.String("us-east-1")
	}

	tagInput := &s3.GetBucketTaggingInput{}
	tagInput.SetBucket(name)
	c, ok := s.RegionClients[*location.LocationConstraint]
	if !ok {
		log.Debugf("Skipping over bucket %s because it is in unknown region %s", name, *location.LocationConstraint)
		return nil, nil
	}

	tags, err := c.GetBucketTagging(tagInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			// Bucket not tagged
			case "NoSuchTagSet": // looks like it is not defined in the s3.Err* codes....
				log.Debugf("Bucket %s has no tags", name)
			default:
				return nil, errors.Wrapf(err, "Error fetching tagset for bucket %s", name)
			}
		}
	}
	for _, tag := range tags.TagSet {
		if tag == nil {
			continue
		}
		bucket.WithTag(tag.Key, tag.Value)
	}

	aclInput := &s3.GetBucketAclInput{}
	aclInput.SetBucket(name)
	acl, err := c.GetBucketAcl(aclInput)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not determnine ACL for %s", name)
	}

	for _, grant := range acl.Grants {
		if grant != nil &&
			grant.Grantee != nil &&
			grant.Grantee.ID != nil &&
			acl != nil &&
			acl.Owner != nil &&
			acl.Owner.ID != nil &&
			(*grant.Grantee.ID != *acl.Owner.ID) {
			bucket.WithLabel(s3LabelIsPublic, aws.String(""))
			break
		}
	}
	return bucket, nil
}
