package runner

import (
	cziAws "github.com/chanzuckerberg/reaper/pkg/aws"
	"github.com/chanzuckerberg/reaper/pkg/config"
	"github.com/chanzuckerberg/reaper/pkg/policy"
	"github.com/hashicorp/go-multierror"
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
func (r *Runner) Run(only []string) ([]policy.Violation, error) {
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

	regions := r.Config.AWSRegions

	var errs *multierror.Error
	var violations []policy.Violation
	for _, p := range policies {
		if len(only) > 0 && !contains(only, p.Name) {
			log.Infof("skipping %s", p.Name)
			continue
		}
		log.Infof("Executing policy: \n%s \n=================", p.String())
		if p.MatchResource(map[string]string{"name": "s3"}) {
			v, err := awsClient.EvalS3(accounts, p)

			errs = multierror.Append(errs, err)

			if v != nil {
				violations = append(violations, v...)
			}
		}

		if p.MatchResource(map[string]string{"name": "ec2_instance"}) {
			log.Infof("Evaluating policy: %s ", p.Name)
			err := awsClient.EvalEc2Instance(accounts, p, regions, func(v policy.Violation) {
				violations = append(violations, v)
			})
			errs = multierror.Append(errs, err)
		}
		if p.MatchResource(map[string]string{"name": "vpc"}) {
			log.Infof("Evaluating policy: %s ", p.Name)
			err := awsClient.EvalVPC(accounts, p, regions, func(v policy.Violation) {
				violations = append(violations, v)
			})
			errs = multierror.Append(errs, err)
		}

		if p.MatchResource(map[string]string{"name": "iam_user"}) {
			log.Infof("Evaluating policy: %s", p.Name)
			v, err := awsClient.EvalIAMUser(accounts, p, regions)

			errs = multierror.Append(errs, err)

			if v != nil {
				violations = append(violations, v...)
			}
		}

		if p.MatchResource(map[string]string{"name": "ebs_volume"}) {
			log.Infof("Evaluating policy: %s", p.Name)
			err := awsClient.EvalEbsVolume(accounts, p, regions, func(v policy.Violation) {
				violations = append(violations, v)
			})
			errs = multierror.Append(errs, err)
		}

		if p.MatchResource(map[string]string{"name": "kms_key"}) {
			log.Infof("Evaluating policy: %s", p.Name)
			err := awsClient.EvalKMSKey(accounts, p, regions, func(v policy.Violation) {
				violations = append(violations, v)
			})
			errs = multierror.Append(errs, err)
		}

	}

	return violations, errs.ErrorOrNil()
}

func contains(haystack []string, needle string) bool {
	for _, a := range haystack {
		if a == needle {
			return true
		}
	}
	return false
}
