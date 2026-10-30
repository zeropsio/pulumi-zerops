package zerops

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/assertpreview"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

func TestYAMLProgram(t *testing.T) {
	if os.Getenv("ZEROPS_TOKEN") == "" || os.Getenv("ZEROPS_SANDBOX") != "dev" {
		t.Skip("requires ZEROPS_TOKEN and ZEROPS_SANDBOX=dev")
	}
	binDir, err := filepath.Abs("../bin")
	if err != nil {
		t.Fatal(err)
	}
	program := pulumitest.NewPulumiTest(t, "test-programs/yaml-basic", opttest.LocalProviderPath("zerops", binDir), opttest.SkipInstall())
	program.Up(t)
	assertpreview.HasNoChanges(t, program.Preview(t))
	program.Destroy(t)
}
