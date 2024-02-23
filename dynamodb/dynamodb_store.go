package dynamodb

import (
	"context"
	"errors"
	"fmt"

	"github.com/hizumisen/go-rate-limiter/core"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
)

type DynamoDbStore[T core.Algorithm] struct {
	client    *dynamodb.Client
	tableName *string
}

func NewDynamoDbStore[T core.Algorithm](
	client *dynamodb.Client,
	tableName string,
) *DynamoDbStore[T] {
	return &DynamoDbStore[T]{
		client:    client,
		tableName: &tableName,
	}
}

var _ core.AlgorithmStorer[*core.TokenBucket] = &DynamoDbStore[*core.TokenBucket]{}

const (
	keyKey      = "rateKey"
	algKey      = "alg"
	sortKey     = "sort"
	expireAtKey = "expireAt"
)

type dynamodbItem[T any] struct {
	Key string `dynamodbav:"rateKey"`
	Alg T      `dynamodbav:"alg"`
}

type TableConfiguration struct {
	AttributeDefinitions []types.AttributeDefinition
	KeySchema            []types.KeySchemaElement
	TTLAttributeName     string
}

func GetTableConfiguration() TableConfiguration {
	return TableConfiguration{
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String(keyKey), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String(keyKey), KeyType: types.KeyTypeHash},
		},
		TTLAttributeName: expireAtKey,
	}
}

func (store *DynamoDbStore[T]) decodeAlg(data map[string]types.AttributeValue) (T, error) {
	var dbItem dynamodbItem[T]

	err := attributevalue.UnmarshalMap(data, &dbItem)
	if err != nil {
		return dbItem.Alg, fmt.Errorf("can't unmarshall dynamodb item to go object: %w", err)
	}

	return dbItem.Alg, nil
}

func (store *DynamoDbStore[T]) Store(
	ctx context.Context,
	key string,
	alg T,
) (T, error) {
	var defaultVal T

	dynamoDbAlg, err := attributevalue.Marshal(alg)
	if err != nil {
		return defaultVal, fmt.Errorf("can't marshall `alg` into dynamodb item: %w", err)
	}

	expireAt, err := attributevalue.Marshal(alg.ExpireAt())
	if err != nil {
		return defaultVal, fmt.Errorf("can't marshall `expireAt` into dynamodb item: %w", err)
	}

	request := dynamodb.UpdateItemInput{
		TableName: store.tableName,
		Key: map[string]types.AttributeValue{
			keyKey: &types.AttributeValueMemberS{Value: key},
		},
		ConditionExpression: aws.String(fmt.Sprintf("attribute_not_exists(%s) or  %s < :sort", sortKey, sortKey)),
		UpdateExpression:    aws.String(fmt.Sprintf("SET %s = :alg, %s = :sort, %s = :expireAt", algKey, sortKey, expireAtKey)),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":alg":      dynamoDbAlg,
			":sort":     &types.AttributeValueMemberS{Value: alg.SortValue()},
			":expireAt": expireAt,
		},
		ReturnValuesOnConditionCheckFailure: types.ReturnValuesOnConditionCheckFailureAllOld,
		ReturnValues:                        types.ReturnValueAllNew,
	}

	var attributes map[string]types.AttributeValue

	result, err := store.client.UpdateItem(ctx, &request)
	if err != nil {
		var errCheck *types.ConditionalCheckFailedException
		if errors.As(err, &errCheck) {
			attributes = errCheck.Item
		} else {
			return defaultVal, fmt.Errorf("can't put item into dynamodb: %w", err)
		}
	}

	if attributes == nil {
		attributes = result.Attributes
	}

	alg, err = store.decodeAlg(attributes)
	if err != nil {
		return defaultVal, err
	}

	return alg, nil
}

func (store DynamoDbStore[T]) Load(ctx context.Context, key string) (*T, error) {
	input := &dynamodb.GetItemInput{
		TableName: store.tableName,
		Key: map[string]types.AttributeValue{
			keyKey: &types.AttributeValueMemberS{Value: key},
		},
		ConsistentRead: aws.Bool(true),
	}

	result, err := store.client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("can't get item from dynamodb: %w", err)
	}

	if result.Item == nil {
		return nil, nil
	}

	alg, err := store.decodeAlg(result.Item)
	if err != nil {
		return nil, err
	}

	return &alg, nil
}
