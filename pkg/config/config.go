package config

import (
	"time"

	"github.com/blend/go-sdk/selector"
	"github.com/chanzuckerberg/aws-tidy/pkg/policy"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// TypeResource describes the type of resource
type TypeResource string

const (
	// TypeResourceS3 is an s3 resource
	TypeResourceS3 TypeResource = "s3"
)

// PolicyConfig is the configuration for a policy
type PolicyConfig struct {
	ResourceSelector string         `yaml:"resource_selector"`
	TagSelector      *string        `yaml:"tag_selector"`
	LabelSelector    *string        `yaml:"label_selector"`
	MaxAge           *time.Duration `yaml:"max_age"`

	NotificationMessage *string `yaml:"notification_message"`
}

// Config is the configuration
type Config struct {
	Policies []PolicyConfig `yaml:"policies"`
}

// GetPolicies gets the policies from a config
func (c *Config) GetPolicies() ([]policy.Policy, error) {
	policies := make([]policy.Policy, len(c.Policies))
	for i, cp := range c.Policies {
		rs, err := selector.Parse(cp.ResourceSelector)
		if err != nil {
			return nil, errors.Wrapf(err, "Invalid selector: %s", cp.ResourceSelector)
		}

		var ls selector.Selector
		if cp.LabelSelector != nil {
			ls, err = selector.Parse(*cp.LabelSelector)
			if err != nil {
				return nil, errors.Wrapf(err, "Invalid selector: %s", *cp.LabelSelector)
			}
		}

		var ts selector.Selector
		if cp.TagSelector != nil {
			ts, err = selector.Parse(*cp.TagSelector)
			if err != nil {
				return nil, errors.Wrapf(err, "Invalid selector: %s", *cp.TagSelector)
			}
		}

		p := policy.Policy{
			ResourceSelector: rs,
			LabelSelector:    ls,
			TagSelector:      ts,
			MaxAge:           cp.MaxAge,
		}
		policies[i] = p
	}
	return policies, nil
}

// NewConfig parses a config
func NewConfig(b []byte) (*Config, error) {
	c := &Config{}
	err := yaml.Unmarshal(b, c)
	return c, errors.Wrap(err, "yaml: could not deserialize config")
}
