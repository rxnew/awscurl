package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func loadConfig(ctx context.Context) (*aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	if e, ok := getEnvConfig(&cfg); ok && e.RoleARN != "" {
		cfg.Credentials = stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), e.RoleARN, func(o *stscreds.AssumeRoleOptions) {
			o.RoleSessionName = e.RoleSessionName
		})
	}

	return &cfg, nil
}

func getEnvConfig(cfg *aws.Config) (*config.EnvConfig, bool) {
	for _, s := range cfg.ConfigSources {
		if c, ok := s.(config.EnvConfig); ok {
			return &c, true
		}
	}
	return nil, false
}
