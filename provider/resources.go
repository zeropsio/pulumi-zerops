// Package zerops bridges the public Zerops Terraform Plugin Framework provider.
package zerops

import (
	"bytes"
	"encoding/json"

	pf "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfbridge"
	"github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge"
	"github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge/tokens"
	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/zeropsio/pulumi-zerops/provider/pkg/version"
	zeropstf "github.com/zeropsio/terraform-provider-zerops/provider"
)

// Provider returns the complete pure Plugin Framework bridge mapping.
func Provider() tfbridge.ProviderInfo {
	python := &tfbridge.PythonInfo{PackageName: "pulumi_zerops", RespectSchemaVersion: true, PythonRequires: ">=3.10"}
	python.PyProject.Enabled = true
	info := tfbridge.ProviderInfo{
		P:    pf.ShimProvider(zeropstf.New(version.Terraform, zeropstf.WithUserAgent("pulumi-zerops/"+version.Version))),
		Name: "zerops", DisplayName: "Zerops", Publisher: "Zerops", ResourcePrefix: "zerops",
		Version: version.Version, TFProviderVersion: version.Terraform,
		Homepage: "https://zerops.io", Repository: "https://github.com/zeropsio/pulumi-zerops",
		GitHubOrg: "zeropsio", License: "MIT", Keywords: []string{"zerops", "category/cloud"},
		Description:       "Manage Zerops projects, services, deployments and networking with Pulumi.",
		PluginDownloadURL: "github://api.github.com/zeropsio/pulumi-zerops",
		MetadataInfo:      tfbridge.NewProviderMetadata(nil),
		Config:            map[string]*tfbridge.SchemaInfo{"token": {Secret: new(true)}},
		DocRules: &tfbridge.DocRuleInfo{
			EditRules: func(defaults []tfbridge.DocsEdit) []tfbridge.DocsEdit {
				return append(defaults,
					tfbridge.DocsEdit{
						Path: "service_deploy.md",
						Edit: func(_ string, content []byte) ([]byte, error) {
							// The standalone upstream example omits provider resolution
							// and references a service declared outside the snippet.
							content = bytes.Replace(content,
								[]byte("resource \"zerops_service_deploy\" \"app\" {"),
								[]byte(`terraform {
  required_providers {
    zerops = { source = "zeropsio/zerops" }
  }
}

variable "service_id" {
  type        = string
  description = "Existing runtime service ID."
}

resource "zerops_service_deploy" "app" {`), 1)
							content = bytes.Replace(content, []byte("service_id = zerops_service.app.id"), []byte("service_id = var.service_id"), 1)
							content = bytes.Replace(content, []byte(`source_dir = "${path.module}/app"`), []byte(`source_dir = "./app"`), 1)
							return content, nil
						},
					},
					tfbridge.DocsEdit{
						Path: "service_subdomain_access.md",
						Edit: func(_ string, content []byte) ([]byte, error) {
							// Relative paths preserve the example inputs without the
							// converter's unsupported Terraform path.module expression.
							content = bytes.Replace(content, []byte(`"${path.module}/app.tar.gz"`), []byte(`"./app.tar.gz"`), 1)
							content = bytes.Replace(content, []byte(`"${path.module}/zerops.yml"`), []byte(`"./zerops.yml"`), 1)
							return content, nil
						},
					})
			},
		},
		Python:            python,
		JavaScript:        &tfbridge.JavaScriptInfo{PackageName: "@zerops/pulumi"},
		Golang:            &tfbridge.GolangInfo{ImportBasePath: "github.com/zeropsio/pulumi-zerops/sdk/go/zerops", ModulePath: "github.com/zeropsio/pulumi-zerops/sdk", GenerateResourceContainerTypes: true, RespectSchemaVersion: true},
		CSharp:            &tfbridge.CSharpInfo{RootNamespace: "Zerops", Namespaces: map[string]string{"zerops": "Pulumi"}, RespectSchemaVersion: true},
		Java:              &tfbridge.JavaInfo{BasePackage: "io.zerops", BuildFiles: "gradle"},
		SchemaPostProcessor: func(spec *schema.PackageSpec) {
			// NuGet identity is independent of the provider's module name.
			// Maven publication fields unsupported by JavaInfo are set by stamp-sdk.py.
			mergeLanguage(spec, "csharp", map[string]any{"packageName": "Zerops.Pulumi"})
		},
	}
	if err := info.ComputeTokens(tokens.SingleModule("zerops_", "index", tokens.MakeStandard("zerops"))); err != nil {
		panic(err)
	}
	return info
}

func mergeLanguage(spec *schema.PackageSpec, language string, values map[string]any) {
	var data map[string]any
	if err := json.Unmarshal(spec.Language[language], &data); err != nil {
		panic(err)
	}
	for key, value := range values {
		data[key] = value
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	spec.Language[language] = encoded
}
