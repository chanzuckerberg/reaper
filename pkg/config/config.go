package config

import (
	"io/ioutil"
	"os"
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

// Duration because I really love writing time un/marshal logic
type Duration time.Duration

// UnmarshalYAML custom unmarshal logic
func (d *Duration) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	var s string
	err := unmarshal(&s)
	if err != nil {
		return errors.Wrap(err, "yaml: Unmarshal error")
	}
	t, err := time.ParseDuration(s)
	if err != nil {
		return errors.Wrapf(err, "Could not parse duration: %s", s)
	}
	*d = Duration(t)
	return nil
}

// Duration returns our custom type as a time.Duration handling nils when needed
func (d *Duration) Duration() *time.Duration {
	if d == nil {
		return nil
	}
	duration := time.Duration(*d)
	return &duration
}

// NotificationConfig is a notification config
type NotificationConfig struct {
	MessageTemplate *string `yaml:"message_template"`
}

// PolicyConfig is the configuration for a policy
type PolicyConfig struct {
	ResourceSelector string  `yaml:"resource_selector"`
	TagSelector      *string `yaml:"tag_selector"`
	LabelSelector    *string `yaml:"label_selector"`
	// MaxAge for this resource
	// If it matches the policy and exceeds MaxAge remediation will be taken.
	MaxAge *Duration `yaml:"max_age"`

	Notifications []NotificationConfig `yaml:"notifications"`
}

// Config is the configuration
type Config struct {
	Policies   []PolicyConfig `yaml:"policies"`
	AWSRegions []string       `yaml:"aws_regions"`
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

		notifications := make([]policy.Notification, len(cp.Notifications))
		for j, n := range cp.Notifications {
			notification := policy.Notification{}
			notification.MessageTemplate = n.MessageTemplate
			notifications[j] = notification
		}

		p := policy.Policy{
			ResourceSelector: rs,
			LabelSelector:    ls,
			TagSelector:      ts,
			MaxAge:           cp.MaxAge.Duration(),
			Notifications:    notifications,
		}
		policies[i] = p
	}
	return policies, nil
}

// FromFile reads a config from a file
func FromFile(fileName string) (*Config, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not open file %s", fileName)
	}
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read config file %s contents", fileName)
	}
	config := &Config{}
	err = yaml.Unmarshal(bytes, config)
	return config, errors.Wrapf(err, "Could not Unmarshal config %s", fileName)
}
