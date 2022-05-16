package terralocacon

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"

	"github.com/testcontainers/testcontainers-go"
)

// AdjustLocalstackTerraformFile changes the localhost ports of localstack.tf file at runtime after container is exposed
// before you run this, you should make sure that you have an existing original localstack.tf file.
// empty string will default to "./local/localstack.tf"
func AdjustLocalstackTerraformFile(ctx context.Context, con testcontainers.Container, localstackFileDir string) (string, error) {
	if strings.Trim(localstackFileDir, " ") == "" {
		localstackFileDir = "./local/localstack.tf"
	}
	if !strings.Contains(localstackFileDir, "localstack.tf") {
		return "", fmt.Errorf("localstack.tf file is not found in a given directory: %s", localstackFileDir)
	}

	rawFile, err := os.ReadFile("./local/localstack.tf")
	if err != nil {
		return "", err
	}

	externalPort, err := con.MappedPort(ctx, "4566")
	if err != nil {
		return "", err
	}

	localhostEndpoint := fmt.Sprintf("localhost:%s", externalPort.Port())
	lines := strings.Split(string(rawFile), "\n")
	for i, line := range strings.Split(string(rawFile), "\n") {
		if strings.Contains(line, "localhost:4566") {
			lines[i] = strings.Replace(line, "localhost:4566", localhostEndpoint, 1)
		}
	}

	localstackTerraformDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}

	modifiedLines := strings.Join(lines, "\n")
	err = os.WriteFile(fmt.Sprintf("%s/localstack.tf", localstackTerraformDir), []byte(modifiedLines), fs.FileMode(os.O_RDWR))
	if err != nil {
		return "", err
	}
	err = os.Chmod(fmt.Sprintf("%s/localstack.tf", localstackTerraformDir), os.ModePerm)

	return localstackTerraformDir, err
}

// NewTerraformOpts returns terraform options for needed terraform commands
func NewTerraformOpts(t *testing.T, terraformDir string, maxRetries int) *terraform.Options {
	t.Helper()
	return terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDir,
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
