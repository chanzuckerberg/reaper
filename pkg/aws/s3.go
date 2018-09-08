package aws

import (
	"github.com/aws/aws-sdk-go/aws"
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

// Delete deletes this bucket
func (s *S3Bucket) Delete() error {
	log.Warnf("Would delete bucket %s", s.name)
	return nil
}

// S3Client is an s3 client
type S3Client struct {
	Svc     s3iface.S3API
	Session *session.Session
}

// NewS3 returns a new s3 client
func NewS3(s *session.Session) S3Client {
	return S3Client{Svc: s3.New(s), Session: s}
}

// Walk walks through all s3 buckets
func (s *S3Client) Walk(p policy.Policy) error {
	input := &s3.ListBucketsInput{}
	output, err := s.Svc.ListBuckets(input)
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
		// nothing to do here
		if entity == nil {
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
	bucket := &S3Bucket{}
	bucket.WithCreatedAt(b.CreationDate)

	locationInput := &s3.GetBucketLocationInput{}
	locationInput.SetBucket(*b.Name)
	location, err := s.Svc.GetBucketLocation(locationInput)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get bucket %s location", *b.Name)
	}
	if location != nil && location.LocationConstraint != nil && s.Session.Config.Region != nil {
		log.Warnf("Bucket %s has location constraint %s", *b.Name, *location.LocationConstraint)
	}
	tagInput := &s3.GetBucketTaggingInput{}
	tagInput.SetBucket(name)
	tags, err := s.Svc.GetBucketTagging(tagInput)
	if err != nil {
		return nil, err
	}
	for _, tag := range tags.TagSet {
		if tag == nil {
			continue
		}
		bucket.WithTag(tag.Key, tag.Value)
	}

	aclInput := &s3.GetBucketAclInput{}
	aclInput.SetBucket(name)

	acl, err := s.Svc.GetBucketAcl(aclInput)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not determnin ACL for %s", name)
	}

	for _, grant := range acl.Grants {
		if grant != nil &&
			grant.Grantee != nil &&
			acl.Owner != nil &&
			acl.Owner.ID != nil &&
			(*grant.Grantee.ID != *acl.Owner.ID) {

			bucket.WithLabel(s3LabelIsPublic, aws.String(""))
			break
		}
	}
	return bucket, nil
}
