package examples

import (
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

func TestNodeJSBasic(t *testing.T) {
	opts := programTestOptions(t, "ts-basic")
	opts.Quick = false
	opts.UseNpm = true
	opts.RunBuild = true
	opts.Config = map[string]string{"appName": fmt.Sprintf("pulacc-ts-%06x", time.Now().UnixNano()&0xffffff)}
	name := opts.Config["appName"]
	t.Cleanup(func() {
		endpoint := os.Getenv("ZEROPS_ENDPOINT")
		organization := os.Getenv("ZEROPS_ORGANIZATION_ID")
		if endpoint == "" || organization == "" {
			t.Error("project absence check requires ZEROPS_ENDPOINT and ZEROPS_ORGANIZATION_ID")
			return
		}
		body, err := json.Marshal(map[string]interface{}{
			"search": []map[string]string{{"name": "clientId", "operator": "eq", "value": organization}},
		})
		if err != nil {
			t.Error(err)
			return
		}
		request, err := http.NewRequest(http.MethodPost, strings.TrimRight(endpoint, "/")+"/api/rest/public/project/search", strings.NewReader(string(body)))
		if err != nil {
			t.Error("constructing project absence request")
			return
		}
		request.Header.Set("Authorization", "Bearer "+os.Getenv("ZEROPS_TOKEN"))
		request.Header.Set("Content-Type", "application/json")
		response, err := (&http.Client{Timeout: 30 * time.Second}).Do(request)
		if err != nil {
			t.Error("project absence API request failed")
			return
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Errorf("project absence API returned HTTP %d", response.StatusCode)
			return
		}
		var result struct {
			Items []struct{ Name string } `json:"items"`
		}
		if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
			t.Error("decoding project absence response")
			return
		}
		for _, project := range result.Items {
			if project.Name == name {
				t.Errorf("project %s remains after destroy", name)
				return
			}
		}
		t.Logf("project %s absent after destroy", name)
	})
	if archive := os.Getenv("ZEROPS_NODEJS_SDK_TGZ"); archive != "" {
		archive, err := filepath.Abs(archive)
		if err != nil {
			t.Fatal(err)
		}
		opts.PrePrepareProject = func(info *engine.Projinfo) error {
			manifest := filepath.Join(info.Root, "package.json")
			data, err := os.ReadFile(manifest)
			if err != nil {
				return err
			}
			var pkg map[string]interface{}
			if err := json.Unmarshal(data, &pkg); err != nil {
				return err
			}
			pkg["dependencies"].(map[string]interface{})["@zerops/pulumi"] = "file:" + archive
			data, err = json.MarshalIndent(pkg, "", "  ")
			if err != nil {
				return err
			}
			return os.WriteFile(manifest, data, 0600)
		}
	}
	validate := func(marker string) func(*testing.T, integration.RuntimeValidationStackInfo) {
		return func(t *testing.T, stack integration.RuntimeValidationStackInfo) {
			url, ok := stack.Outputs["url"].(string)
			if !ok || url == "" {
				t.Fatal("missing public URL")
			}
			client := &http.Client{Timeout: 20 * time.Second}
			deadline := time.Now().Add(10 * time.Minute)
			for time.Now().Before(deadline) {
				response, err := client.Get(url)
				if err == nil {
					body, readErr := io.ReadAll(response.Body)
					response.Body.Close()
					if readErr == nil && response.StatusCode == 200 && strings.Contains(string(body), marker) {
						t.Logf("HTTP 200 contains %s", marker)
						return
					}
				}
				time.Sleep(10 * time.Second)
			}
			t.Fatalf("app did not return HTTP 200 with %s", marker)
		}
	}
	edit := t.TempDir()
	if err := os.MkdirAll(filepath.Join(edit, "app"), 0700); err != nil {
		t.Fatal(err)
	}
	server, err := os.ReadFile("ts-basic/app/server.js")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(edit, "app/server.js"), []byte(strings.ReplaceAll(string(server), "v1", "v2")), 0600); err != nil {
		t.Fatal(err)
	}
	opts.ExtraRuntimeValidation = validate("(v1)")
	opts.EditDirs = []integration.EditDir{{Dir: edit, Additive: true, ExtraRuntimeValidation: validate("(v2)")}}
	integration.ProgramTest(t, &opts)
}

func TestTemplateNodeJS(t *testing.T) {
	opts := programTestOptions(t, "")
	template, err := filepath.Abs("../templates/zerops-typescript")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	command := exec.Command("pulumi", "new", template, "--dir", directory, "--generate-only", "--yes", "--name", "zerops-ts-template", "--description", "Zerops TypeScript starter")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("instantiate template: %v\n%s", err, output)
	}
	opts.Dir = directory
	opts.Quick = false
	opts.UseNpm = true
	opts.RunBuild = true
	opts.Config = map[string]string{
		"appName": fmt.Sprintf("pulacc-tpl-ts-%06x", time.Now().UnixNano()&0xffffff),
		"database": "true",
	}
	componentPlugin, err := filepath.Abs("../bin")
	if err != nil {
		t.Fatal(err)
	}
	opts.LocalProviders = append(opts.LocalProviders, integration.LocalDependency{Package: "zerops-app", Path: componentPlugin})
	localPackages := make(map[string]string)
	for name, sdk := range map[string]string{
		"@zerops/pulumi": "../sdk/nodejs/bin",
		"@zerops/pulumi-app": "../sdk-app/nodejs/bin",
	} {
		sdk, err := filepath.Abs(sdk)
		if err != nil {
			t.Fatal(err)
		}
		destination := t.TempDir()
		command := exec.Command("npm", "pack", "--json", "--pack-destination", destination)
		command.Dir = sdk
		output, err := command.Output()
		if err != nil {
			t.Fatalf("pack local %s: %v", name, err)
		}
		var packages []struct {
			Filename string `json:"filename"`
		}
		if err := json.Unmarshal(output, &packages); err != nil {
			t.Fatal(err)
		}
		if len(packages) != 1 {
			t.Fatalf("pack local %s returned %d packages", name, len(packages))
		}
		localPackages[name] = filepath.Join(destination, packages[0].Filename)
	}
	opts.PrePrepareProject = func(info *engine.Projinfo) error {
		manifest := filepath.Join(info.Root, "package.json")
		data, err := os.ReadFile(manifest)
		if err != nil {
			return err
		}
		var pkg map[string]interface{}
		if err := json.Unmarshal(data, &pkg); err != nil {
			return err
		}
		dependencies := pkg["dependencies"].(map[string]interface{})
		for name, variable := range map[string]string{
			"@zerops/pulumi": "ZEROPS_NODEJS_SDK_TGZ",
			"@zerops/pulumi-app": "ZEROPS_NODEJS_APP_SDK_TGZ",
		} {
			dependencies[name] = "file:" + localPackages[name]
			if archive := os.Getenv(variable); archive != "" {
				archive, err = filepath.Abs(archive)
				if err != nil {
					return err
				}
				dependencies[name] = "file:" + archive
			}
		}
		data, err = json.MarshalIndent(pkg, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(manifest, data, 0600)
	}
	validate := func(marker string) func(*testing.T, integration.RuntimeValidationStackInfo) {
		return func(t *testing.T, stack integration.RuntimeValidationStackInfo) {
			url, ok := stack.Outputs["url"].(string)
			if !ok || url == "" {
				t.Fatal("missing public URL")
			}
			client := &http.Client{Timeout: 20 * time.Second}
			deadline := time.Now().Add(10 * time.Minute)
			for time.Now().Before(deadline) {
				response, err := client.Get(url)
				if err == nil {
					body, readErr := io.ReadAll(response.Body)
					response.Body.Close()
					if readErr == nil && response.StatusCode == 200 &&
						strings.Contains(string(body), marker) &&
						strings.Contains(string(body), "DATABASE_URL set: true") {
						t.Logf("HTTP 200 contains %s and DATABASE_URL set: true", marker)
						return
					}
				}
				time.Sleep(10 * time.Second)
			}
			t.Fatal("app did not return expected body with database configured")
		}
	}
	edit := t.TempDir()
	if err := os.MkdirAll(filepath.Join(edit, "app"), 0700); err != nil {
		t.Fatal(err)
	}
	server, err := os.ReadFile(filepath.Join(directory, "app/server.py"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(edit, "app/server.py"), []byte(strings.ReplaceAll(string(server), "Hello from Zerops WebApp!", "Updated Zerops WebApp!")), 0600); err != nil {
		t.Fatal(err)
	}
	opts.ExtraRuntimeValidation = validate("Hello from Zerops WebApp!")
	opts.EditDirs = []integration.EditDir{{Dir: edit, Additive: true, ExtraRuntimeValidation: validate("Updated Zerops WebApp!")}}
	integration.ProgramTest(t, &opts)
}
