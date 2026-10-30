package examples

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

// programTestOptions shares credential guards and the local plugin with each
// language's examples_<language>_test.go; callers can add language dependencies.
func programTestOptions(t *testing.T, directory string) integration.ProgramTestOptions {
	t.Helper()
	if os.Getenv("ZEROPS_TOKEN") == "" || os.Getenv("ZEROPS_SANDBOX") != "dev" {
		t.Skip("requires ZEROPS_TOKEN and ZEROPS_SANDBOX=dev")
	}
	plugin, err := filepath.Abs("../bin")
	if err != nil {
		t.Fatal(err)
	}
	return integration.ProgramTestOptions{
		Dir:            directory,
		Quick:          true,
		SkipRefresh:    true,
		Verbose:        true,
		LocalProviders: []integration.LocalDependency{{Package: "zerops", Path: plugin}},
	}
}
