package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/chanzuckerberg/reaper/pkg/policy"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// s3 specific labels
const (
	s3LabelACLPublic     TypeEntityLabel = "s3_acl_public"
	s3LabelACLPublicRead TypeEntityLabel = "s3_acl_public_read"
)

// S3Bucket is an evaluation entity representing an s3 bucket
type S3Bucket struct {
	Entity
	name string
}

// NewS3Bucket returns a new s3 bucket entity
func NewS3Bucket(name string) *S3Bucket {
	bucket := &S3Bucket{
		Entity: NewEntity(),
		name:   name,
	}
	bucket.ID = name
	return bucket
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

// GetConsoleURL will return a URL for this resource in the AWS console
func (s *S3Bucket) GetConsoleURL() string {
	t := "https://s3.console.aws.amazon.com/s3/buckets/%s/"
	return fmt.Sprintf(t, s.ID)
}

// EvalS3 walks through all s3 buckets
func (c *Client) EvalS3(accounts []*policy.Account, p policy.Policy) ([]policy.Violation, error) {
	log.Infof("Walking s3 buckets")
	var violations []policy.Violation
	var errs error
	ctx := context.Background()
	for _, account := range accounts {
		log.Infof("walking account %s (%d)", account.Name, account.ID)
		listOutput, err := c.Get(account.ID, account.Role, account.ExternalID, DefaultRegion).S3.ListBuckets(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "Could not list buckets")
		}
		for _, bucket := range listOutput.Buckets {
			res, err := c.DescribeS3Bucket(account.ID, account.Role, account.ExternalID, bucket)
			// accumulate errors
			if err != nil {
				errs = multierror.Append(errs, err)
				continue
			}
			if res == nil {
				log.Debugf("Nil bucket - nothing to do")
				continue
			}
			if p.Match(res) {
				violation := policy.NewViolation(p, res, false, account)
				violations = append(violations, violation)
			}

		}
	}

	return violations, errs
}

// DescribeS3Bucket describes the bucket
func (c *Client) DescribeS3Bucket(accountID int64, roleName string, externalID string, b *s3.Bucket) (*S3Bucket, error) {
	ctx := context.Background()
	if b.Name == nil {
		return nil, errors.New("Nil bucket name")
	}
	log.Debugf("Describing bucket %s", *b.Name)
	name := *b.Name
	bucket := NewS3Bucket(name)
	bucket.AddCreatedAt(b.CreationDate)

	location, err := c.Get(accountID, roleName, externalID, DefaultRegion).S3.GetBucketLocation(ctx, name)
	if err != nil {
		return nil, err
	}
	bucket.Region = location

	tagInput := &s3.GetBucketTaggingInput{}
	tagInput.SetBucket(name)
	regionalClient := c.Get(accountID, roleName, externalID, location)
	if regionalClient == nil {
		log.Debugf("Skipping over bucket %s because it is in unknown region %s", name, location)
		return nil, nil
	}

	tags, err := regionalClient.S3.GetBucketTagging(ctx, name)
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
		bucket.AddTag(tag.Key, tag.Value)
	}

	acl, err := regionalClient.S3.GetBucketACL(ctx, name)
	if err != nil {
		return nil, err
	}

	log.Debugf("reading grants for bucket %s", name)
	for _, grant := range acl.Grants {

		log.Debugf("grant %#v", grant)

		if grant != nil &&
			grant.Grantee != nil &&
			grant.Grantee.Type != nil &&
			*grant.Grantee.Type == "Group" &&
			grant.Grantee.URI != nil &&
			*grant.Grantee.URI == "http://acs.amazonaws.com/groups/global/AllUsers" {

			log.Debugf("public bucket %s", name)
			bucket.AddLabel(s3LabelACLPublic, aws.String(""))

			if grant.Permission != nil &&
				*grant.Permission == "READ" {
				bucket.AddLabel(s3LabelACLPublicRead, aws.String(""))
			}
		}
	}
	return bucket, nil
}
