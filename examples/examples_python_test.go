package examples

import (
	"crypto/rand"
	"encoding/hex"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestPythonBasic(t *testing.T) {
	options := programTestOptions(t, "python-basic")
	var suffix [3]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		t.Fatal(err)
	}
	wheelDirectory, err := filepath.Abs("../sdk/python/dist")
	if err != nil {
		t.Fatal(err)
	}
	// Resolve the published requirement from the local wheel before linking the
	// source SDK. Build it first with make build_python.
	options.Env = []string{"PIP_FIND_LINKS=" + wheelDirectory}
	options.Dependencies = []string{"../sdk/python"}
	options.Config = map[string]string{"appName": "pulacc-python-" + hex.EncodeToString(suffix[:])}
	integration.ProgramTest(t, &options)
}

func TestTemplatePython(t *testing.T) {
	options := programTestOptions(t, "../templates/zerops-python")
	template, err := filepath.Abs(options.Dir)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	command := exec.Command("pulumi", "new", template, "--dir", directory,
		"--generate-only", "--yes", "--name", "zerops-python-template",
		"--description", "A Python WebApp template integration test")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("instantiate Python template: %v\n%s", err, output)
	}
	var suffix [3]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		t.Fatal(err)
	}
	primitiveWheels, err := filepath.Abs("../sdk/python/dist")
	if err != nil {
		t.Fatal(err)
	}
	componentWheels, err := filepath.Abs("../sdk-app/python/dist")
	if err != nil {
		t.Fatal(err)
	}
	options.Dir = directory
	options.Env = []string{"PIP_FIND_LINKS=" + primitiveWheels + " " + componentWheels}
	options.Dependencies = []string{"../sdk/python", "../sdk-app/python"}
	options.LocalProviders = append(options.LocalProviders, integration.LocalDependency{
		Package: "zerops-app", Path: options.LocalProviders[0].Path,
	})
	options.Config = map[string]string{
		"appName":  "pulacc-tpl-python-" + hex.EncodeToString(suffix[:]),
		"database": "true",
	}
	options.ExtraRuntimeValidation = goBasicHTTP("DATABASE_URL set: true")
	integration.ProgramTest(t, &options)
}
