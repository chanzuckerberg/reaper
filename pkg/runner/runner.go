package runner

import (
	cziAws "github.com/chanzuckerberg/aws-tidy/pkg/aws"
	"github.com/chanzuckerberg/aws-tidy/pkg/config"
	"github.com/chanzuckerberg/aws-tidy/pkg/policy"
	log "github.com/sirupsen/logrus"
)

// Runner takes a config and generates all the violations
type Runner struct {
	Config *config.Config
}

// New will construct a Runner object with the given Config
func New(c *config.Config) *Runner {
	return &Runner{Config: c}
}

// Run will evaluate all the polices against the accounts in the config and return violations
func (r *Runner) Run() ([]*policy.Violation, error) {

	policies, err := r.Config.GetPolicies()
	if err != nil {
		return nil, err
	}

	accounts, err := r.Config.GetAccounts()
	if err != nil {
		return nil, nil
	}

	awsClient, err := cziAws.NewClient(accounts, r.Config.AWSRegions)
	if err != nil {
		return nil, err
	}

	var violations []*policy.Violation
	for _, p := range policies {
		log.Infof("Executing policy: \n%s \n=================", p.String())
		if p.MatchResource(map[string]string{"name": "s3"}) {
			v, err := awsClient.EvalS3(accounts, &p)
			if err != nil {
				return nil, err
			}
			if v != nil {
				violations = append(violations, v...)
			}
		}
	}

	return violations, err
}
