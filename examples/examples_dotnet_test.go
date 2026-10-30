package examples

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pulumi/pulumi/pkg/v3/engine"
	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestCSharpBasic(t *testing.T) {
	opts := programTestOptions(t, "csharp-basic")
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, ".cache", "nuget-local")
	pack := exec.Command("dotnet", "pack", filepath.Join(root, "sdk", "dotnet"), "-c", "Release", "-o", source, "-p:PackageVersion=0.1.1")
	if output, err := pack.CombinedOutput(); err != nil {
		t.Fatalf("pack SDK: %v\n%s", err, output)
	}
	opts.PrepareProject = func(project *engine.Projinfo) error {
		dir, _, err := project.GetPwdMain()
		if err != nil {
			return err
		}
		cmd := exec.Command("dotnet", "restore", "--source", source, "--source", "https://api.nuget.org/v3/index.json")
		cmd.Dir = dir
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("restore: %w\n%s", err, output)
		}
		return nil
	}
	var suffix [3]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		t.Fatal(err)
	}
	opts.Config = map[string]string{"appName": "pulacc-csharp-" + hex.EncodeToString(suffix[:])}
	opts.Quick = false
	opts.SkipExportImport = true
	opts.NoParallel = true
	validate := func(marker string) func(*testing.T, integration.RuntimeValidationStackInfo) {
		return func(t *testing.T, stack integration.RuntimeValidationStackInfo) {
			url, ok := stack.Outputs["url"].(string)
			if !ok || url == "" {
				t.Fatal("missing URL output")
			}
			client := &http.Client{Timeout: 20 * time.Second}
			deadline := time.Now().Add(10 * time.Minute)
			for time.Now().Before(deadline) {
				response, err := client.Get(url)
				if err == nil {
					body, readErr := io.ReadAll(response.Body)
					response.Body.Close()
					if readErr == nil && response.StatusCode == http.StatusOK && strings.Contains(string(body), marker) {
						t.Logf("HTTP 200 with %s", marker)
						return
					}
				}
				time.Sleep(5 * time.Second)
			}
			t.Fatalf("application did not return HTTP 200 with %s", marker)
		}
	}
	opts.ExtraRuntimeValidation = validate("(v1)")
	edit := t.TempDir()
	if err := os.Mkdir(filepath.Join(edit, "app"), 0755); err != nil {
		t.Fatal(err)
	}
	server, err := os.ReadFile("csharp-basic/app/server.js")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(edit, "app", "server.js"), []byte(strings.ReplaceAll(string(server), `"v1"`, `"v2"`)), 0644); err != nil {
		t.Fatal(err)
	}
	opts.EditDirs = []integration.EditDir{{Dir: edit, Additive: true, ExtraRuntimeValidation: validate("(v2)")}}
	integration.ProgramTest(t, &opts)
	body, err := json.Marshal(map[string]any{"search": []map[string]string{
		{"name": "clientId", "operator": "eq", "value": os.Getenv("ZEROPS_ORGANIZATION_ID")},
	}})
	if err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequest(http.MethodPost, strings.TrimRight(os.Getenv("ZEROPS_ENDPOINT"), "/")+"/api/rest/public/project/search", strings.NewReader(string(body)))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+os.Getenv("ZEROPS_TOKEN"))
	request.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{Timeout: 30 * time.Second}).Do(request)
	if err != nil {
		t.Fatal("project absence query failed")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("project absence query returned HTTP %d", response.StatusCode)
	}
	var projects struct {
		Items []struct {
			Name string `json:"name"`
		} `json:"items"`
	}
	if err := json.NewDecoder(response.Body).Decode(&projects); err != nil {
		t.Fatal("invalid project search response")
	}
	for _, project := range projects.Items {
		if project.Name == opts.Config["appName"] {
			t.Fatal("project remains after destroy")
		}
	}
	t.Log("project absent after destroy")
}

func TestTemplateDotnet(t *testing.T) {
	opts := programTestOptions(t, "")
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	cmd := exec.Command("pulumi", "new", filepath.Join(root, "templates", "zerops-csharp"),
		"--dir", directory, "--name", "zerops-csharp-template-test", "--description", "CSharp starter acceptance",
		"--generate-only", "--yes")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("instantiate template: %v\n%s", err, output)
	}
	source := filepath.Join(root, ".cache", "nuget-local")
	for _, sdk := range []string{"sdk", "sdk-app"} {
		cmd := exec.Command("dotnet", "pack", filepath.Join(root, sdk, "dotnet"), "-c", "Release", "-o", source, "-p:PackageVersion=0.1.1")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("pack %s: %v\n%s", sdk, err, output)
		}
	}
	opts.Dir = directory
	opts.PrepareProject = func(project *engine.Projinfo) error {
		dir, _, err := project.GetPwdMain()
		if err != nil {
			return err
		}
		cmd := exec.Command("dotnet", "restore", "--source", source, "--source", "https://api.nuget.org/v3/index.json")
		cmd.Dir = dir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("restore: %w\n%s", err, output)
		}
		return nil
	}
	var suffix [3]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		t.Fatal(err)
	}
	opts.Config = map[string]string{"appName": "pulacc-tpl-csharp-" + hex.EncodeToString(suffix[:]), "database": "true"}
	opts.Quick = false
	opts.SkipExportImport = true
	opts.NoParallel = true
	componentPlugin, err := filepath.Abs("../bin")
	if err != nil {
		t.Fatal(err)
	}
	opts.LocalProviders = append(opts.LocalProviders, integration.LocalDependency{Package: "zerops-app", Path: componentPlugin})
	expectedBody := "Hello from Zerops WebApp!"
	opts.ExtraRuntimeValidation = func(t *testing.T, stack integration.RuntimeValidationStackInfo) {
		url, ok := stack.Outputs["url"].(string)
		if !ok || url == "" {
			t.Fatal("missing URL output")
		}
		client := &http.Client{Timeout: 20 * time.Second}
		deadline := time.Now().Add(10 * time.Minute)
		for time.Now().Before(deadline) {
			response, err := client.Get(url)
			if err == nil {
				body, readErr := io.ReadAll(response.Body)
				response.Body.Close()
				if readErr == nil && response.StatusCode == http.StatusOK && strings.Contains(string(body), "DATABASE_URL set: true") && strings.Contains(string(body), expectedBody) {
					t.Log("HTTP 200 with DATABASE_URL set: true")
					return
				}
			}
			time.Sleep(5 * time.Second)
		}
		t.Fatal("application did not return HTTP 200 with database connection")
	}
	edit := t.TempDir()
	if err := os.Mkdir(filepath.Join(edit, "app"), 0755); err != nil {
		t.Fatal(err)
	}
	server, err := os.ReadFile(filepath.Join(directory, "app", "server.py"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(edit, "app", "server.py"), []byte(strings.ReplaceAll(string(server), "Hello from Zerops WebApp!", "Hello from Zerops WebApp v2!")), 0644); err != nil {
		t.Fatal(err)
	}
	validate := opts.ExtraRuntimeValidation
	opts.EditDirs = []integration.EditDir{{Dir: edit, Additive: true, ExtraRuntimeValidation: func(t *testing.T, stack integration.RuntimeValidationStackInfo) {
		expectedBody = "Hello from Zerops WebApp v2!"
		validate(t, stack)
	}}}
	integration.ProgramTest(t, &opts)
}
