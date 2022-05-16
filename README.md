# Terralocacon

# What is this?

Terralocacon is a utility kit for debugging ***a single lambda*** with localstack running AWS infrastructure inside a container. It allows you to write a simple terraform file for localstack and apply it to a container so that you can debug ***your single lambda*** in your own machine

# What is the use case then?

The use case is you have production data for the input(preferably JSON) of your lambda, then you have databases or you have buckets or you have HTTP/SOAP/FTP clients, etc... which depends on your single lambda. And you have no more clear vision in your mocks(your mocks can not cover complex issues like stale data in your database or HTTP client-related timeout throttling). That's where you will be eligible to use this and hopefully solve your integration issue without doing live debugging on actual AWS infra or mocking hundred of resources which you are not sure about its data quality.

# Dependencies

- Localstack (the docker image)
- Testcontainers (a library for creating containers)
- Docker connections (a library for mapping container ports)
- Terratest (a library for applying with terraform)

# Getting Started

## 0- Get that

```bash
docker pull localstack/localstack@latest
go get -u github.com/xiatechs/terralocacon
```

## 1- Define your infra resources for localstack (localstack.tf)

```bash
region                      = "eu-west-1"
  access_key                  = "fake"
  secret_key                  = "fake"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  s3_force_path_style         = true

  endpoints {
    dynamodb = "http://localhost:4566"
    kinesis  = "http://localhost:4566"
    s3       = "http://localhost:4566"
  }
}

resource "aws_dynamodb_table" "sales_order_table" {
  name           = "salesorder"
  read_capacity  = "20"
  write_capacity = "20"
  hash_key       = "PK"
  range_key      = "SK"

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }
}

resource "aws_kinesis_stream" "localstack_processed_sales_order_events_kinesis" {
  name             = "localstack-processed-sales-order-events"
  shard_count      = 1
  retention_period = 30

  shard_level_metrics = [
    "IncomingBytes",
    "OutgoingBytes",
  ]
}

resource "aws_s3_bucket" "localstack_archiving_bucket" {
  bucket = "localstack-archiving-bucket"
}
```

## 2- Swap out AWS session in your lambda (main.go)

```go
var sess *session.Session
if strings.TrimSpace(os.Getenv("LOCALSTACK_ENDPOINT")) != "" {
	sess, _ = session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("REGION_AWS")),
		Endpoint:    aws.String(fmt.Sprintf("http://%s", os.Getenv("LOCALSTACK_ENDPOINT"))),
		Credentials: credentials.NewStaticCredentials("fake", "fake", ""),
	})
} else {
	sess = session.Must(session.NewSession())
}
```

## 3- Write your test (main_test.go)

```go
package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xiatechs/terralocacon"
)

const awsRegion = "eu-west-1"
const awsServices = "dynamodb,kinesis,s3"

func Test_run(t *testing.T) {
	ctx := context.Background()
	localstackContainer, port, err := terralocacon.NewLocalstackContainer(ctx, awsRegion, awsServices)
	require.NoError(t, err)
	defer func() {
		if err = terralocacon.TerminateContainer(ctx, localstackContainer); err != nil {
			t.Logf("failed to tear down localstack container: %v", err)
		}
	}()
	localstackTerraformDir, err := terralocacon.AdjustLocalstackTerraformFile(ctx, localstackContainer, "./local/localstack.tf")
	require.NoError(t, err)
	opts := terralocacon.NewTerraformOpts(t, localstackTerraformDir, 0)
	terralocacon.Apply(t, opts)
	defer terralocacon.Destroy(t, opts)

	// mandatory env variables
	_ = os.Setenv("LOCALSTACK_ENDPOINT", fmt.Sprintf("localhost:%s", port))
	_ = os.Setenv("REGION_AWS", awsRegion)
	// set your own env variables here for debugging

	// wrap your test data and execute
	var ev events.KinesisEvent
	ev.Records = make([]events.KinesisEventRecord, 1)
	ev.Records[0] = events.KinesisEventRecord{
		Kinesis: events.KinesisRecord{
			Data: []byte(`{"test":"test"}`),
		},
	}
	Init()
	defer Shutdown()
	err = Run(ev)
	assert.Nil(t, err)
}
```
