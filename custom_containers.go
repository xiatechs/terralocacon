package terralocacon

import (
	"context"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// NewContainer creates a new container with the given name and image.
func NewContainer(ctx context.Context, req *testcontainers.ContainerRequest) (testcontainers.Container, error) {
	if strings.TrimSpace(req.Image) == "" {
		return nil, fmt.Errorf("didn't specify any image nmae for the container")
	}

	if len(req.ExposedPorts) == 0 {
		return nil, fmt.Errorf("didn't specify any port to expose for the container")
	}

	con, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        req.Image,
			ExposedPorts: req.ExposedPorts,
			WaitingFor:   req.WaitingFor,
			Env:          req.Env,
		},
		Started: true,
	})

	return con, err
}

// TerminateContainer terminates a given container
func TerminateContainer(ctx context.Context, con testcontainers.Container) error {
	return con.Terminate(ctx)
}

// NewLocalstackContainer creates a new localstack container
func NewLocalstackContainer(ctx context.Context, awsRegion, awsServices string) (testcontainers.Container, string, error) {
	conRequest := &testcontainers.ContainerRequest{
		Image:        "localstack/localstack:latest",
		ExposedPorts: []string{"4566/tcp"},
		WaitingFor:   wait.ForLog("Ready"),
		Env: map[string]string{
			"DEFAULT_REGION":   awsRegion,
			"SERVICES":         awsServices,
			"KINESIS_PROVIDER": "kinesalite",
			"DOCKER_HOST":      "unix:///var/run/docker.sock",
			"DATA_DIR":         "/tmp/localstack/data",
			"DEBUG":            "1",
		},
	}

	c, err := NewContainer(ctx, conRequest)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create localstack container: %s", err.Error())
	}

	externalPort, err := c.MappedPort(ctx, "4566")
	if err != nil {
		return nil, "", fmt.Errorf("failed to get the external port of created container: %s", err.Error())
	}

	return c, externalPort.Port(), nil
}

// NewMongoDBContainer creates a new mongo container
func NewMongoDBContainer(ctx context.Context, username, password string) (testcontainers.Container, string, error) {
	conRequest := &testcontainers.ContainerRequest{
		Image:        "mongo:latest",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("ready for start up."),
		Env: map[string]string{
			"MONGO_INITDB_ROOT_USERNAME": username,
			"MONGO_INITDB_ROOT_PASSWORD": password,
		},
	}

	c, err := NewContainer(ctx, conRequest)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create mongo container: %s", err.Error())
	}

	externalPort, err := c.MappedPort(ctx, "27017")
	if err != nil {
		return nil, "", fmt.Errorf("failed to get the external port of created container: %s", err.Error())
	}

	return c, externalPort.Port(), nil
}
