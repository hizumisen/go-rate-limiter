package benchmark

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hizumisen/go-rate-limiter/core"
	"github.com/hizumisen/go-rate-limiter/dynamodb"
)

func inmemoryStore[T core.Algorithm]() *core.InMmemoryStore[T] {
	return core.NewInMemoryStore[T](50)
}

func dynamodbStore[T core.Algorithm]() *dynamodb.DynamoDbStore[T] {
	ctx := context.Background()
	tableConfig := dynamodb.GetTableConfiguration()
	tableName := fmt.Sprintf("rate-limit-%d", time.Now().UnixNano())

	client, err := dynamodb.NewDynamodbClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = dynamodb.CreateTableIfMissing(ctx, client, tableName, tableConfig)
	if err != nil {
		log.Fatal(err)
	}

	return dynamodb.NewDynamoDbStore[T](client, tableName)
}
