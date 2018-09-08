package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cziAws "github.com/chanzuckerberg/aws-tidy/pkg/aws"
	"github.com/chanzuckerberg/aws-tidy/pkg/config"
)

// main
func main() {
	conf := config.Config{
		Policies: []config.PolicyConfig{
			config.PolicyConfig{
				ResourceSelector: "name", // All resources
				TagSelector:      aws.String("managedBy"),
				LabelSelector:    aws.String(""),
			},
		},
		Regions: []string{
			"us-east-1",
			"us-east-2",
			"us-west-1",
			"us-west-2",
		},
	}

	policies, err := conf.GetPolicies()
	if err != nil {
		panic(err)
	}

	s := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	awsClient, err := cziAws.NewClient(s, conf.Regions)
	if err != nil {
		panic(err)
	}

	for _, p := range policies {
		if p.MatchResource(map[string]string{"name": "s3"}) {
			err := awsClient.S3.Walk(p)
			if err != nil {
				panic(err)
			}
		}
	}
}
