package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func NewAwsLocalConfig(ctx context.Context) (aws.Config, error) {
	credential := aws.Credentials{
		AccessKeyID:     "dummy",
		SecretAccessKey: "dummy",
		SessionToken:    "dummy",
	}

	awsConfig, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("eu-west-1"),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(_, _ string, _ ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: "http://localhost:8000"}, nil
			},
		)),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: credential,
		}),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("can't create aws cofiguration: %w", err)
	}

	return awsConfig, nil
}
func NewDynamodbClient(ctx context.Context) (*dynamodb.Client, error) {
	awsConfig, err := NewAwsLocalConfig(ctx)
	if err != nil {
		return nil, err
	}

	return dynamodb.NewFromConfig(awsConfig), nil
}

func CreateTableIfMissing(
	ctx context.Context,
	client *dynamodb.Client,
	tableName string,
	config TableConfiguration,
) error {
	table, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: &tableName})
	if err != nil {
		var notFoundErr *types.ResourceNotFoundException
		if !errors.As(err, &notFoundErr) {
			return fmt.Errorf("can't describe table: %w", err)
		}
	}

	if table != nil {
		if !reflect.DeepEqual(table.Table.KeySchema, config.KeySchema) {
			return fmt.Errorf("invalid key schema, found %v", table.Table.KeySchema)
		}
		if !reflect.DeepEqual(table.Table.AttributeDefinitions, config.AttributeDefinitions) {
			return fmt.Errorf("invalid attribute definitions, found %v", table.Table.AttributeDefinitions)
		}

		return nil
	}

	_, err = client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName:            &tableName,
		AttributeDefinitions: config.AttributeDefinitions,
		KeySchema:            config.KeySchema,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
	})
	if err != nil {
		return fmt.Errorf("can't create table: %w", err)
	}

	err = dynamodb.NewTableExistsWaiter(client).
		Wait(
			ctx,
			&dynamodb.DescribeTableInput{TableName: &tableName},
			10*time.Second,
		)
	if err != nil {
		return fmt.Errorf("can't wait for table creation: %w", err)
	}

	_, err = client.UpdateTimeToLive(ctx, &dynamodb.UpdateTimeToLiveInput{
		TableName: &tableName,
		TimeToLiveSpecification: &types.TimeToLiveSpecification{
			AttributeName: &config.TTLAttributeName,
			Enabled:       aws.Bool(true),
		},
	})
	if err != nil {
		return fmt.Errorf("can't set time to live table: %w", err)
	}

	return nil
}
