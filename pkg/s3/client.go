package s3

import (
	"ai-learn-english/config"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"

	s3_config "github.com/aws/aws-sdk-go-v2/config"
	s3_credentials "github.com/aws/aws-sdk-go-v2/credentials"
	s3_provider "github.com/aws/aws-sdk-go-v2/service/s3"
)

func GetClient() (*s3_provider.Client, error) {
	// Build AWS config for MinIO (S3-compatible)
	s3cfg := config.Cfg.S3
	region := s3cfg.Region
	if region == "" {
		region = "us-east-1"
	}

	opts := []func(*s3_config.LoadOptions) error{
		s3_config.WithRegion(region),
	}
	if s3cfg.AccessKey != "" && s3cfg.SecretKey != "" {
		opts = append(opts, s3_config.WithCredentialsProvider(
			s3_credentials.NewStaticCredentialsProvider(
				s3cfg.AccessKey,
				s3cfg.SecretKey,
				"",
			),
		))
	}

	cfg, err := s3_config.LoadDefaultConfig(
		context.TODO(),
		opts...,
	)
	if err != nil {
		return nil, err
	}

	endpoint := s3cfg.Endpoint
	client := s3_provider.NewFromConfig(cfg, func(o *s3_provider.Options) {
		o.UsePathStyle = true
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint) // e.g., http://localhost:9000
		}
	})
	return client, nil
}

func GetPresignClient() (*s3_provider.PresignClient, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}
	return s3_provider.NewPresignClient(client), nil
}
