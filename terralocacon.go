package terralocacon

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// NewContainer creates a new container with the given name and image.
func NewContainer(ctx context.Context, image string, ports []string, env map[string]string) (testcontainers.Container, error) {
	con, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        image,
			ExposedPorts: ports,
			WaitingFor:   wait.ForLog("Ready."),
			Env:          env,
		},
		Started: true,
	})
	return con, err
}

// NewLocalstackContainer creates a new localstack container
func NewLocalstackContainer(ctx context.Context, awsRegion string, awsServices string) (testcontainers.Container, error) {
	conRequest := testcontainers.ContainerRequest{
		Image:        "localstack/localstack:latest",
		ExposedPorts: []string{"4566/tcp"},
		WaitingFor:   wait.ForLog("Ready."),
		Env: map[string]string{
			"DEFAULT_REGION":   awsRegion,
			"SERVICES":         awsServices,
			"KINESIS_PROVIDER": "kinesalite",
			"DOCKER_HOST":      "unix:///var/run/docker.sock",
			"DATA_DIR":         "/tmp/localstack/data",
			"DEBUG":            "1",
		},
	}
	return NewContainer(ctx, conRequest.Image, conRequest.ExposedPorts, conRequest.Env)
}

// TerminateContainer terminates a given container
func TerminateContainer(ctx context.Context, con testcontainers.Container) error {
	return con.Terminate(ctx)
}

// MakeTerraformTempDir changes the localhost ports of localstack.tf file at runtime after container is exposed
// before you run this, you should make sure that you have an existing original localstack.tf file.
// empty string will default to "./local/localstack.tf"
func MakeTerraformTempDir(ctx context.Context, con testcontainers.Container, localstackFileDir string) (string, string, error) {
	if strings.Trim(localstackFileDir, " ") == "" {
		localstackFileDir = "./local/localstack.tf"
	}
	if !strings.Contains(localstackFileDir, "localstack.tf") {
		return "", "", fmt.Errorf("localstack.tf file is not found in a given directory: %s", localstackFileDir)
	}

	rawFile, err := os.ReadFile("./local/localstack.tf")
	if err != nil {
		return "", "", err
	}

	internalPort, err := nat.NewPort("tcp", "4566")
	if err != nil {
		return "", "", err
	}

	externalPort, err := con.MappedPort(ctx, internalPort)
	if err != nil {
		return "", "", err
	}

	localhostEndpoint := fmt.Sprintf("localhost:%s", externalPort.Port())
	lines := strings.Split(string(rawFile), "\n")
	for i, line := range strings.Split(string(rawFile), "\n") {
		if strings.Contains(line, "localhost:4566") {
			lines[i] = strings.Replace(line, "localhost:4566", localhostEndpoint, 1)
		}
	}

	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", "", err
	}

	modifiedLines := strings.Join(lines, "\n")
	err = os.WriteFile(fmt.Sprintf("%s/localstack.tf", tempDir), []byte(modifiedLines), fs.FileMode(os.O_RDWR))
	if err != nil {
		return "", "", err
	}

	return tempDir, localhostEndpoint, err
}

// NewTerraformOpts returns terraform options for needed terraform commands
func NewTerraformOpts(t *testing.T, tempTerraformDir string, maxRetries int) *terraform.Options {
	t.Helper()
	return terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: tempTerraformDir,
		MaxRetries:   maxRetries,
	})
}

// Apply runs init and apply commands with terraform
func Apply(t *testing.T, opts *terraform.Options) {
	t.Helper()
	terraform.InitAndApply(t, opts)
}

// Destroy destroys the underlying infrastructure with terraform
func Destroy(t *testing.T, opts *terraform.Options) {
	t.Helper()
	terraform.Destroy(t, opts)
}
