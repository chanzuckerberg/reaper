package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/chanzuckerberg/aws-tidy/pkg/policy"
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

// S3Client is an s3 client
type S3Client struct {
	Client s3iface.S3API
	// Per region clients
	RegionClients map[string]s3iface.S3API
	Session       *session.Session
}

// NewS3 returns a new s3 client
func NewS3(s *session.Session, regions []string) *S3Client {
	s3Client := &S3Client{
		Client:        s3.New(s),
		Session:       s,
		RegionClients: map[string]s3iface.S3API{},
	}
	for _, region := range regions {
		s3Client.RegionClients[region] = s3.New(s, &aws.Config{Region: aws.String(region)})
	}
	return s3Client
}

// Walk walks through all s3 buckets
func (s *S3Client) Walk(p policy.Policy) error {
	input := &s3.ListBucketsInput{}
	output, err := s.Client.ListBuckets(input)
	if err != nil {
		return errors.Wrap(err, "Could not list buckets")
	}

	for _, bucket := range output.Buckets {
		if bucket == nil {
			continue
		}
		log.Infof("Considering bucket %s", *bucket.Name)
		entity, err := s.DescribeBucket(bucket)
		if err != nil {
			return err
		}
		if entity == nil {
			log.Infof("Skipping bucket %s", *bucket.Name)
			continue
		}
		if p.Match(entity) {
			log.Infof("Matched bucket %s", *bucket.Name)
		}
	}
	return nil
}

// DescribeBucket describes this bucket
func (s *S3Client) DescribeBucket(b *s3.Bucket) (*S3Bucket, error) {
	if b.Name == nil {
		return nil, errors.New("Nil bucket name")
	}
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
		log.Infof("Skipping over bucket %s because it is in unknown region %s", name, *location.LocationConstraint)
		return nil, nil
	}

	tags, err := c.GetBucketTagging(tagInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NoSuchTagSet": // looks like it is not defined in the s3.Err* codes....
				log.Infof("Bucket %s has no tags", name)
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
