package examples

import (
	"crypto/rand"
	"encoding/hex"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestYAMLBasic(t *testing.T) {
	options := programTestOptions(t, "yaml-basic")
	integration.ProgramTest(t, &options)
}

func TestTemplateYAML(t *testing.T) {
	options := programTestOptions(t, "")
	directory := t.TempDir()
	template, err := filepath.Abs("../templates/zerops-yaml")
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("pulumi", "new", template, "--dir", directory,
		"--name", "zerops-yaml-template-test", "--description", "YAML starter integration test",
		"--generate-only", "--yes", "--non-interactive")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("instantiate YAML template: %v\n%s", err, output)
	}
	var suffix [3]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		t.Fatal(err)
	}
	options.Dir = directory
	options.Config = map[string]string{
		"appName": "pulacc-tpl-yaml-" + hex.EncodeToString(suffix[:]),
		"database": "true",
	}
	options.LocalProviders = append(options.LocalProviders, integration.LocalDependency{
		Package: "zerops-app",
		Path: options.LocalProviders[0].Path,
	})
	options.ExtraRuntimeValidation = func(t *testing.T, stack integration.RuntimeValidationStackInfo) {
		for _, key := range []string{"url", "projectId", "serviceId"} {
			if value, ok := stack.Outputs[key].(string); !ok || value == "" {
				t.Errorf("missing WebApp output %s", key)
			}
		}
		goBasicHTTP("DATABASE_URL set: true")(t, stack)
	}
	integration.ProgramTest(t, &options)
}
