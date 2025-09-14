package urls

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoRepo struct {
	client    *dynamodb.Client
	tableName string
	log       *slog.Logger
}

// NewDynamoRepo creates a Dynamo-backed repository. Use the same endpoint logic as readiness.
func NewDynamoRepo(endpoint, tableName string, logger *slog.Logger) *DynamoRepo {
	var cfg aws.Config
	var err error

	if endpoint != "" {
		resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               endpoint,
				PartitionID:       "aws",
				SigningRegion:     "us-east-1",
				HostnameImmutable: true,
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

	return &DynamoRepo{
		client:    dynamodb.NewFromConfig(cfg),
		tableName: tableName,
		log:       logger,
	}
}

func (r *DynamoRepo) CreateURL(ctx context.Context, item URLItem) error {
	// TODO: Use PutItem with ConditionExpression attribute_not_exists(PK).
	return ErrNotImplemented
}

func (r *DynamoRepo) GetURL(ctx context.Context, shortCode string) (URLItem, error) {
	// TODO: GetItem by PK=URL#{shortCode}, SK=URL; map to URLItem.
	return URLItem{}, ErrNotImplemented
}

func (r *DynamoRepo) IncrementClicks(ctx context.Context, shortCode string) (int64, error) {
	// TODO: UpdateItem ADD clicks :one; return new value; NotFound -> ErrNotFound.
	return 0, ErrNotImplemented
}
