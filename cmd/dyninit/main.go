package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func main() {
	endpoint := getenv("DYNAMO_ENDPOINT", "http://localhost:8000")
	table := getenv("URLS_TABLE", "urls")

	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               endpoint,
			PartitionID:       "aws",
			SigningRegion:     "us-east-1",
			HostnameImmutable: true,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("local", "local", "")),
	)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	cli := dynamodb.NewFromConfig(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err = cli.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(table)})
	if err == nil {
		fmt.Printf("Table %s already exists.\n", table)
		return
	}
	var rnfe *types.ResourceNotFoundException
	if !errors.As(err, &rnfe) {
		log.Fatalf("describe table failed: %v", err)
	}

	_, err = cli.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(table),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("PK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("SK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("GSI1PK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("GSI1SK"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("PK"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("SK"), KeyType: types.KeyTypeRange},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("GSI1"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("GSI1PK"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("GSI1SK"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	})
	if err != nil {
		log.Fatalf("create table failed: %v", err)
	}

	deadline := time.Now().Add(20 * time.Second)
	for {
		time.Sleep(500 * time.Millisecond)
		out, err := cli.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(table)})
		if err == nil && out.Table != nil && out.Table.TableStatus == types.TableStatusActive {
			break
		}
		if time.Now().After(deadline) {
			log.Fatalf("table %s not ACTIVE in time", table)
		}
	}
	fmt.Printf("Table %s created and ACTIVE.\n", table)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
