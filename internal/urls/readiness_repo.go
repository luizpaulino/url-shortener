package urls

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type ReadinessRepo interface {
	Ping(ctx context.Context) error
}

type dynamoReadinessRepo struct {
	client    *dynamodb.Client
	tableName string
	log       *slog.Logger
}

func NewDynamoReadinessRepo(endpoint, tableName string, logger *slog.Logger) ReadinessRepo {
	var cfg aws.Config
	var err error

	if endpoint != "" {
		// Local DynamoDB config with static dummy credentials
		resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               endpoint,
				PartitionID:       "aws",
				SigningRegion:     "us-east-1",
				HostnameImmutable: true, // evita reescrita do host
			}, nil
		})

		cfg, err = config.LoadDefaultConfig(
			context.Background(),
			config.WithRegion("us-east-1"),
			config.WithEndpointResolverWithOptions(resolver),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("local", "local", "")),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(context.Background())
	}

	if err != nil {
		logger.Error("dynamo_config_error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	return &dynamoReadinessRepo{
		client:    dynamodb.NewFromConfig(cfg),
		tableName: tableName,
		log:       logger,
	}
}

func (r *dynamoReadinessRepo) Ping(ctx context.Context) error {
	_, err := r.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(r.tableName),
	})
	if err != nil {
		return fmt.Errorf("describe_table_failed: %w", err)
	}
	return nil
}
