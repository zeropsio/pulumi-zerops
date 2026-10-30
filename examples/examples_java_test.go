package examples

import (
	"crypto/rand"
	"fmt"
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

func TestJavaBasic(t *testing.T) {
	options := programTestOptions(t, "java-basic")
	name := make([]byte, 3)
	if _, err := rand.Read(name); err != nil {
		t.Fatal(err)
	}
	options.Config = map[string]string{"appName": fmt.Sprintf("pulacc-java-%x", name)}
	// The test honors an optional Maven local repository passed via MAVEN_ARGS.
	// Publish sdk/java to that repository before invoking this test.
	if mavenArgs := os.Getenv("MAVEN_ARGS"); mavenArgs != "" {
		options.Env = append(options.Env, "MAVEN_ARGS="+mavenArgs)
	}
	options.Quick = false
	options.ExtraRuntimeValidation = javaHTTPBody("(v1)")
	options.SkipExportImport = true
	update := t.TempDir()
	if err := os.CopyFS(update, os.DirFS("java-basic")); err != nil {
		t.Fatal(err)
	}
	server := filepath.Join(update, "app", "server.js")
	body, err := os.ReadFile(server)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(server, []byte(strings.ReplaceAll(string(body), "v1", "v2")), 0600); err != nil {
		t.Fatal(err)
	}
	options.EditDirs = []integration.EditDir{{Dir: update, Additive: true, ExtraRuntimeValidation: javaHTTPBody("(v2)")}}
	integration.ProgramTest(t, &options)
}

func javaHTTPBody(marker string) func(*testing.T, integration.RuntimeValidationStackInfo) {
	return func(t *testing.T, stack integration.RuntimeValidationStackInfo) {
		t.Helper()
		url, ok := stack.Outputs["url"].(string)
		if !ok || url == "" {
			t.Fatal("missing url output")
		}
		client := &http.Client{Timeout: 20 * time.Second}
		deadline := time.Now().Add(8 * time.Minute)
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
		t.Fatalf("application did not serve HTTP 200 with %s", marker)
	}
}

func TestTemplateJava(t *testing.T) {
	options := programTestOptions(t, "")
	name := make([]byte, 3)
	if _, err := rand.Read(name); err != nil {
		t.Fatal(err)
	}
	appName := fmt.Sprintf("pulacc-tpl-java-%x", name)
	template, err := filepath.Abs("../templates/zerops-java")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	command := exec.Command("pulumi", "new", template, "--dir", directory,
		"--generate-only", "--yes", "--name", appName, "--stack", "dev",
		"--config", "appName="+appName)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("instantiate Java template: %v\n%s", err, output)
	}
	repository, err := filepath.Abs("../.cache/m2")
	if err != nil {
		t.Fatal(err)
	}
	options.Dir = directory
	options.Env = append(options.Env, "MAVEN_ARGS=-Dmaven.repo.local="+repository)
	options.Config = map[string]string{"appName": appName, "database": "true"}
	options.LocalProviders = append(options.LocalProviders,
		integration.LocalDependency{Package: "zerops-app", Path: options.LocalProviders[0].Path})
	options.Quick = false
	options.SkipExportImport = true
	options.ExtraRuntimeValidation = javaHTTPBody("WebApp! DATABASE_URL set: true")
	update := t.TempDir()
	if err := os.CopyFS(update, os.DirFS(directory)); err != nil {
		t.Fatal(err)
	}
	server := filepath.Join(update, "app", "server.py")
	body, err := os.ReadFile(server)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(server, []byte(strings.ReplaceAll(string(body), "WebApp!", "WebApp v2!")), 0600); err != nil {
		t.Fatal(err)
	}
	options.EditDirs = []integration.EditDir{{Dir: update, Additive: true,
		ExtraRuntimeValidation: javaHTTPBody("WebApp v2! DATABASE_URL set: true")}}
	integration.ProgramTest(t, &options)
}
