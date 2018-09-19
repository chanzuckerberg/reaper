package runner

import (
	"github.com/aws/aws-sdk-go/aws/session"
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

	awsClient, err := awsClient(r.Config.AWSRegions)
	if err != nil {
		return nil, err
	}

	var violations []*policy.Violation
	for _, p := range policies {
		log.Infof("Executing policy: \n%s \n=================", p.String())
		if p.MatchResource(map[string]string{"name": "s3"}) {
			v, err := awsClient.EvalS3(&p)
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

func awsClient(regions []string) (*cziAws.Client, error) {
	s := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return cziAws.NewClient(s, regions)
}
