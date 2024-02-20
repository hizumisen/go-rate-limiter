module github.com/hizumisen/go-rate-limiter/internal

go 1.21.6

replace github.com/hizumisen/go-rate-limiter/core => ../core

replace github.com/hizumisen/go-rate-limiter/dynamodb => ../dynamodb

require (
	github.com/hizumisen/go-rate-limiter/core v0.0.0-00010101000000-000000000000
	github.com/hizumisen/go-rate-limiter/dynamodb v0.0.0-00010101000000-000000000000
	github.com/wcharczuk/go-chart/v2 v2.1.1
	golang.org/x/sync v0.6.0
)

require (
	github.com/aws/aws-sdk-go v1.50.21 // indirect
	github.com/aws/aws-sdk-go-v2 v1.25.0 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.27.0 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.13.2 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.15.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.29.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.19.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.9.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.19.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.22.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.27.0 // indirect
	github.com/aws/smithy-go v1.20.0 // indirect
	github.com/blend/go-sdk v1.20220411.3 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	golang.org/x/image v0.11.0 // indirect
)
