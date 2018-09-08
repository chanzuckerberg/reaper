package aws

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
)

// Client is an AWS client
type Client struct {
	KMS *KMSClient
	S3  *S3Client
}

// Entity is an AWS entity s3 bucket, ec2 instance, etc
type Entity struct {
	// Labels are resource specific properties
	labels map[string]string
	// Tags are AWS tags
	tags map[string]string

	// createdAt for this object
	createdAt *time.Time
}

// common labels
const (
	labelID  = "id"
	labelARN = "arn"
)

// TypeEntityLabel An EntityLabel
type TypeEntityLabel string

// NewEntity returns a new aws entity
func NewEntity() Entity {
	return Entity{
		labels:    map[string]string{},
		tags:      map[string]string{},
		createdAt: nil,
	}
}

// Delete deletes
func (e *Entity) Delete() error {
	return errors.New("Delete not implemented")
}

// GetLabels returns this entitie's labels
func (e *Entity) GetLabels() map[string]string {
	return e.labels
}

// GetTags returns the tags
func (e *Entity) GetTags() map[string]string {
	return e.tags
}

// GetCreatedAt returns createdAt
func (e *Entity) GetCreatedAt() *time.Time {
	return e.createdAt
}

// WithLabel adds a label if the value is not nil
func (e *Entity) WithLabel(key TypeEntityLabel, value *string) *Entity {
	if e.labels == nil {
		e.labels = map[string]string{}
	}
	if value != nil {
		e.labels[string(key)] = *value
	}
	return e
}

// WithTag adds a tag if the value is not nill
func (e *Entity) WithTag(key *string, value *string) *Entity {
	if e.tags == nil {
		e.tags = map[string]string{}
	}
	if value != nil && key != nil {
		e.tags[*key] = *value
	}
	return e
}

// WithCreatedAt adds a createdAt
func (e *Entity) WithCreatedAt(t *time.Time) *Entity {
	e.createdAt = t
	return e
}

// WalkFun is a walk function over AWS entities
type WalkFun func(*Entity, error) error

// NewClient returns a new aws client
func NewClient(sess *session.Session, regions []string) (*Client, error) {
	client := &Client{
		KMS: NewKMS(sess),
		S3:  NewS3(sess, regions),
	}
	return client, nil
}
