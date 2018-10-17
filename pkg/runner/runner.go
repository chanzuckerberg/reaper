package runner

import (
	"github.com/aws/aws-sdk-go/service/support"
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

	r.UpdateTrustedAdvisorChecks(awsClient, accounts)

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
		if p.MatchResource(map[string]string{"name": "ec2_security_group"}) {
			log.Infof("Evaluating policy: %s", p.Name)
			err := awsClient.EvalEC2SG(accounts, p, regions, func(v policy.Violation) {
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

		if p.MatchResource(map[string]string{"name": "iam_access_key"}) {
			log.Infof("Evaluating policy: %s", p.Name)
			v, err := awsClient.EvalIAMAccessKey(accounts, p)
			errs = multierror.Append(errs, err)
			if v != nil {
				violations = append(violations, v...)
			}
		}

	}

	return violations, errs.ErrorOrNil()
}

// UpdateTrustedAdvisorChecks will walk all accounts, all checks and update them if they are stale
func (r *Runner) UpdateTrustedAdvisorChecks(client *cziAws.Client, accounts []*policy.Account) {
	for _, a := range accounts {
		log.Debugf("ta for %s", a.Name)
		c := client.Get(a.ID, a.Role, "us-east-1")
		en := "en"
		input := &support.DescribeTrustedAdvisorChecksInput{
			Language: &en,
		}
		output, _ := c.Support.Svc.DescribeTrustedAdvisorChecks(input)
		checkIds := []*string{}
		for _, check := range output.Checks {
			// log.Debugf("check: %#v", check)
			checkIds = append(checkIds, check.Id)
		}
		refreshStatusInput := &support.DescribeTrustedAdvisorCheckRefreshStatusesInput{
			CheckIds: checkIds,
		}
		out, _ := c.Support.Svc.DescribeTrustedAdvisorCheckRefreshStatuses(refreshStatusInput)
		for _, status := range out.Statuses {
			log.Debug("status: %#v", status)
			if status.MillisUntilNextRefreshable != nil && *status.MillisUntilNextRefreshable == 0 {
				log.Debugf("refreshing %s", *status.CheckId)
				refreshInput := &support.RefreshTrustedAdvisorCheckInput{
					CheckId: status.CheckId,
				}
				c.Support.Svc.RefreshTrustedAdvisorCheck(refreshInput)
				// TODO wait until no longer pending
			}
		}
	}
}

func contains(haystack []string, needle string) bool {
	for _, a := range haystack {
		if a == needle {
			return true
		}
	}
	return false
}
