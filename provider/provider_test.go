package zerops

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
)

func TestTokens(t *testing.T) {
	info := Provider()
	resources := []string{"IntegrationToken", "ObjectStorage", "Project", "ProjectDedicatedIpv4", "ProjectEnv", "ProjectImport", "ProjectL7BalancerConfig", "ProjectPrometheusConfig", "ProjectRemoteLogging", "ProjectVpnPeer", "PublicHttpRouting", "PublicPortRouting", "Service", "ServiceBackupConfig", "ServiceDeploy", "ServiceEnv", "ServiceGitIntegration", "ServiceRestart", "ServiceSubdomainAccess"}
	functions := []string{"Locations", "Organization", "Project", "ProjectExport", "Projects", "Regions", "Service", "ServiceEnvVars", "ServiceTypes"}
	if len(info.Resources) != 19 || len(info.DataSources) != 9 {
		t.Fatalf("unexpected mappings: %d resources, %d data sources", len(info.Resources), len(info.DataSources))
	}
	for _, name := range resources {
		token := "zerops:index/" + strings.ToLower(name[:1]) + name[1:] + ":" + name
		found := false
		for _, mapping := range info.Resources {
			if string(mapping.Tok) == token {
				found = true
			}
		}
		if !found {
			t.Errorf("missing resource token %s", token)
		}
	}
	for _, name := range functions {
		token := "zerops:index/get" + name + ":get" + name
		found := false
		for _, mapping := range info.DataSources {
			if string(mapping.Tok) == token {
				found = true
			}
		}
		if !found {
			t.Errorf("missing function token %s", token)
		}
	}
	if info.Config["token"].Secret == nil || !*info.Config["token"].Secret {
		t.Fatal("token must be secret")
	}
}

func TestCommittedSchema(t *testing.T) {
	content, err := os.ReadFile("cmd/pulumi-resource-zerops/schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var spec schema.PackageSpec
	if err := json.Unmarshal(content, &spec); err != nil {
		t.Fatal(err)
	}
	if len(spec.Resources) != 19 {
		t.Fatalf("schema has %d resources", len(spec.Resources))
	}
	if !spec.Config.Variables["token"].Secret {
		t.Fatal("schema token is not secret")
	}
	for _, key := range []string{"token", "endpoint", "organizationId", "postgresqlMinDiskGb", "routingForceSync"} {
		property, exists := spec.Config.Variables[key]
		if !exists || property.Description == "" {
			t.Errorf("config %s missing or undocumented", key)
		}
	}
}

func TestSchemaSnapshot(t *testing.T) {
	// Build a fresh generator so this also catches drift in mappings/upstream.
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("make", "check_schema")
	command.Dir = root
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("schema snapshot: %v\n%s", err, output)
	}
}
