package aws

import (
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"
)

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

// GetOwner will return this entity's owner as indicated by the 'owner' tag.
func (e *Entity) GetOwner() string {
	if e.tags != nil {
		o, ok := e.tags["owner"]
		if ok {
			return o
		}
	}
	return ""
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

// WithBoolLabel adds a label if the value is true
func (e *Entity) WithBoolLabel(key TypeEntityLabel, value *bool) *Entity {
	if value != nil && *value {
		return e.WithLabel(key, aws.String(""))
	}
	return e
}

// WithInt64Label adds a label if the value is not nil
func (e *Entity) WithInt64Label(key TypeEntityLabel, value *int64) *Entity {
	if value != nil {
		return e.WithLabel(key, aws.String(strconv.FormatInt(*value, 10)))
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
