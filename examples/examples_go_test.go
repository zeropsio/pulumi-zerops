package examples

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestGoBasic(t *testing.T) {
	options := programTestOptions(t, "go-basic")
	options.StackName = "dev"
	options.NoParallel = true
	options.Verbose = true
	suffix := make([]byte, 3)
	if _, err := rand.Read(suffix); err != nil {
		t.Fatal(err)
	}
	options.Config = map[string]string{"appName": "pulacc-go-" + hex.EncodeToString(suffix)}
	// ProgramTest copies the project into a temporary directory, so replace the
	// example's relative SDK path with an absolute path in that copy only.
	sdk, err := filepath.Abs("../sdk")
	if err != nil {
		t.Fatal(err)
	}
	options.Dependencies = []string{sdk}
	options.Quick = false
	options.SkipExportImport = true
	options.ExtraRuntimeValidation = goBasicHTTP("(v1)")
	edit := t.TempDir()
	if err := os.Mkdir(filepath.Join(edit, "app"), 0o755); err != nil {
		t.Fatal(err)
	}
	server, err := os.ReadFile("go-basic/app/server.js")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(server), `"v1"`) {
		t.Fatal("app source is missing the v1 version marker")
	}
	if err := os.WriteFile(filepath.Join(edit, "app", "server.js"), []byte(strings.ReplaceAll(string(server), `"v1"`, `"v2"`)), 0o600); err != nil {
		t.Fatal(err)
	}
	options.EditDirs = []integration.EditDir{{
		Dir: edit, Additive: true, ExtraRuntimeValidation: goBasicHTTP("(v2)"),
	}}
	integration.ProgramTest(t, &options)
	goBasicProjectAbsent(t, options.Config["appName"])
}

func goBasicHTTP(marker string) func(*testing.T, integration.RuntimeValidationStackInfo) {
	return func(t *testing.T, stack integration.RuntimeValidationStackInfo) {
		t.Helper()
		url, ok := stack.Outputs["url"].(string)
		if !ok || url == "" {
			t.Fatal("missing public url output")
		}
		client := &http.Client{Timeout: 20 * time.Second}
		deadline := time.Now().Add(8 * time.Minute)
		for time.Now().Before(deadline) {
			response, err := client.Get(url)
			if err == nil {
				body, readErr := io.ReadAll(io.LimitReader(response.Body, 4096))
				response.Body.Close()
				if readErr == nil && response.StatusCode == http.StatusOK && strings.Contains(string(body), marker) {
					t.Logf("HTTP 200 with %s", marker)
					return
				}
			}
			time.Sleep(10 * time.Second)
		}
		t.Fatalf("public app did not return HTTP 200 with %s", marker)
	}
}

func goBasicProjectAbsent(t *testing.T, appName string) {
	t.Helper()
	endpoint := os.Getenv("ZEROPS_ENDPOINT")
	organization := os.Getenv("ZEROPS_ORGANIZATION_ID")
	if endpoint == "" || organization == "" {
		t.Fatal("cleanup verification requires ZEROPS_ENDPOINT and ZEROPS_ORGANIZATION_ID")
	}
	body, err := json.Marshal(map[string]any{"search": []map[string]string{{
		"name": "clientId", "operator": "eq", "value": organization,
	}}})
	if err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequest(http.MethodPost, strings.TrimRight(endpoint, "/")+"/api/rest/public/project/search", strings.NewReader(string(body)))
	if err != nil {
		t.Fatal("invalid cleanup API endpoint")
	}
	request.Header.Set("Authorization", "Bearer "+os.Getenv("ZEROPS_TOKEN"))
	request.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{Timeout: time.Minute}).Do(request)
	if err != nil {
		t.Fatal("cleanup API request failed")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("cleanup API returned HTTP %d", response.StatusCode)
	}
	var result struct {
		Items []struct {
			Name string `json:"name"`
		} `json:"items"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatal("invalid cleanup API response")
	}
	if result.Items == nil {
		t.Fatal("cleanup API response has no items array")
	}
	for _, project := range result.Items {
		if project.Name == appName {
			t.Fatal("example project remains after destroy")
		}
	}
	t.Log("cleanup API HTTP 200: project absent")
}

func TestTemplateGo(t *testing.T) {
	options := programTestOptions(t, "")
	options.StackName = "dev"
	options.NoParallel = true
	options.Verbose = true
	suffix := make([]byte, 3)
	if _, err := rand.Read(suffix); err != nil {
		t.Fatal(err)
	}
	appName := "pulacc-tpl-go-" + hex.EncodeToString(suffix)
	template, err := filepath.Abs("../templates/zerops-go")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	command := exec.Command("pulumi", "new", template, "--generate-only", "--yes",
		"--name", "zerops-go-starter", "--description", "Go starter integration test", "--dir", directory)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("instantiate Go starter: %v\n%s", err, output)
	}
	options.Dir = directory
	options.Config = map[string]string{"appName": appName, "database": "true"}
	for _, module := range []string{"sdk", "sdk-app"} {
		path, err := filepath.Abs("../" + module)
		if err != nil {
			t.Fatal(err)
		}
		options.Dependencies = append(options.Dependencies, path)
	}
	plugin, err := filepath.Abs("../bin")
	if err != nil {
		t.Fatal(err)
	}
	options.LocalProviders = append(options.LocalProviders, integration.LocalDependency{
		Package: "zerops-app", Path: plugin,
	})
	options.Quick = false
	options.SkipExportImport = true
	options.RunBuild = true
	options.ExtraRuntimeValidation = goBasicHTTP("Hello from Zerops WebApp! DATABASE_URL set: true")
	edit := t.TempDir()
	if err := os.Mkdir(filepath.Join(edit, "app"), 0o755); err != nil {
		t.Fatal(err)
	}
	server, err := os.ReadFile(filepath.Join(directory, "app", "server.py"))
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.ReplaceAll(string(server), "Hello from Zerops WebApp!", "Hello from Zerops WebApp v2!")
	if updated == string(server) {
		t.Fatal("template app source is missing the expected update marker")
	}
	if err := os.WriteFile(filepath.Join(edit, "app", "server.py"), []byte(updated), 0o600); err != nil {
		t.Fatal(err)
	}
	options.EditDirs = []integration.EditDir{{
		Dir: edit, Additive: true,
		ExtraRuntimeValidation: goBasicHTTP("Hello from Zerops WebApp v2! DATABASE_URL set: true"),
	}}
	integration.ProgramTest(t, &options)
	goBasicProjectAbsent(t, appName)
}
