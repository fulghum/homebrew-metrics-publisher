package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// LoadParameter loads an encrypted parameter from AWS SSM Parameter Store.
// We use Parameter Store instead of Secrets Manager since each secret costs $.40 a month to store.
func LoadParameter(parameterName string) (*string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		// Optionally use a shared config profile for local development
		//config.WithSharedConfigProfile("jason+test@dolthub.com"),
		// TODO: Make region configurable from infrastructure code
		config.WithRegion("us-west-2"))
	if err != nil {
		return nil, err
	}

	// Create an SSM client
	client := ssm.NewFromConfig(cfg)

	output, err := client.GetParameter(context.TODO(), &ssm.GetParameterInput{
		Name:           &parameterName,
		WithDecryption: true,
	})
	if err != nil {
		return nil, err
	}

	return output.Parameter.Value, nil
}
