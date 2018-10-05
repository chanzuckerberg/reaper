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
	ID        string
	Name      string
	Region    string
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

// GetRegion returns the region in which this entity exists
func (e *Entity) GetRegion() string {
	return e.Region
}

// GetLabels returns this entitie's labels
func (e *Entity) GetLabels() map[string]string {
	return e.labels
}

//GetLabelOr will return the label value (if defined). otherwise `or`. Useful for templates.
func (e *Entity) GetLabelOr(label string, or string) string {
	l, ok := e.labels[label]
	if ok {
		return l
	}
	return or
}

// GetTags returns the tags
func (e *Entity) GetTags() map[string]string {
	return e.tags
}

// GetCreatedAt returns createdAt
func (e *Entity) GetCreatedAt() *time.Time {
	return e.createdAt
}

// GetName returns a user-friendly string identifying the Entity
func (e *Entity) GetName() string {
	return e.Name
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

// AddLabel adds a label if the value is not nil
func (e *Entity) AddLabel(key TypeEntityLabel, value *string) *Entity {
	if e.labels == nil {
		e.labels = map[string]string{}
	}
	if value != nil {
		e.labels[string(key)] = *value
	}
	return e
}

// AddBoolLabel adds a label if the value is true
func (e *Entity) AddBoolLabel(key TypeEntityLabel, value *bool) *Entity {
	if value != nil && *value {
		return e.AddLabel(key, aws.String("true"))
	}
	return e
}

// AddInt64Label adds a label if the value is not nil
func (e *Entity) AddInt64Label(key TypeEntityLabel, value *int64) *Entity {
	if value != nil {
		return e.AddLabel(key, aws.String(strconv.FormatInt(*value, 10)))
	}
	return e
}

// AddTag adds a tag if the value is not nill
func (e *Entity) AddTag(key *string, value *string) *Entity {
	if e.tags == nil {
		e.tags = map[string]string{}
	}
	if value != nil && key != nil {
		e.tags[*key] = *value
	}
	return e
}

// AddCreatedAt adds a createdAt
func (e *Entity) AddCreatedAt(t *time.Time) *Entity {
	e.createdAt = t
	return e
}
