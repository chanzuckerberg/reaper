package config

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/chanzuckerberg/reaper/pkg/policy"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/labels"
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
	Recipient       string `yaml:"recipient"`
	MessageTemplate string `yaml:"message_template"`
}

// PolicyConfig is the configuration for a policy
type PolicyConfig struct {
	Name             string  `yaml:"name"`
	ResourceSelector string  `yaml:"resource_selector"`
	TagSelector      *string `yaml:"tag_selector"`
	LabelSelector    *string `yaml:"label_selector"`
	// MaxAge for this resource
	// If it matches the policy and exceeds MaxAge remediation will be taken.
	MaxAge *Duration `yaml:"max_age"`

	Notifications []NotificationConfig `yaml:"notifications"`
}

//AccountConfig identifies an AWS account we want to monitor
type AccountConfig struct {
	Name  string `yaml:"name"`
	ID    int64  `yaml:"id"`
	Role  string `yaml:"role"`
	Owner string `yaml:"owner"`
}

//IdentityMapConfig will allow mapping group email lists to slack channels
type IdentityMapConfig struct {
	Email string `yaml:"email"`
	Slack string `yaml:"slack"`
}

// Config is the configuration
type Config struct {
	Version     int                 `yaml:version`
	Policies    []PolicyConfig      `yaml:"policies"`
	AWSRegions  []string            `yaml:"aws_regions"`
	Accounts    []AccountConfig     `yaml:"accounts"`
	IdentityMap []IdentityMapConfig `yaml:"identity_map"`
}

// GetPolicies gets the policies from a config
func (c *Config) GetPolicies() ([]policy.Policy, error) {
	policies := make([]policy.Policy, len(c.Policies))
	for i, cp := range c.Policies {
		rs, err := labels.Parse(cp.ResourceSelector)
		if err != nil {
			return nil, errors.Wrapf(err, "Invalid selector: %s", cp.ResourceSelector)
		}

		var ls labels.Selector
		if cp.LabelSelector != nil {
			ls, err = labels.Parse(*cp.LabelSelector)
			if err != nil {
				return nil, errors.Wrapf(err, "Invalid selector: %s", *cp.LabelSelector)
			}
		}

		var ts labels.Selector
		if cp.TagSelector != nil {
			ts, err = labels.Parse(*cp.TagSelector)
			if err != nil {
				return nil, errors.Wrapf(err, "Invalid selector: %s", *cp.TagSelector)
			}
		}

		notifications := make([]policy.Notification, len(cp.Notifications))
		for j, n := range cp.Notifications {
			notification := policy.Notification{}
			notification.MessageTemplate = n.MessageTemplate
			notification.Recipient = n.Recipient
			notifications[j] = notification
		}

		p := policy.Policy{
			Name:             cp.Name,
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

//GetAccounts will return policy.Account objects
func (c *Config) GetAccounts() ([]*policy.Account, error) {
	var accounts []*policy.Account
	for _, a := range c.Accounts {
		accounts = append(accounts, &policy.Account{Name: a.Name, ID: a.ID, Role: a.Role, Owner: a.Owner})
	}
	return accounts, nil
}

// GetIdentityMap will return a map of email -> slack identifier
func (c *Config) GetIdentityMap() (map[string]string, error) {
	m := make(map[string]string)
	for _, i := range c.IdentityMap {
		m[i.Email] = i.Slack
	}
	return m, nil
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
