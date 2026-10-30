package main

import (
 "context"
 "encoding/json"
 "fmt"
 "os"

 "github.com/blang/semver"
 "github.com/pulumi/pulumi/pkg/v3/codegen/schema"
 provider "github.com/pulumi/pulumi-go-provider"
 "github.com/pulumi/pulumi-go-provider/infer"
 components "github.com/zeropsio/pulumi-zerops/components"
)

var version = "0.1.1"

func main() {
 p, err := infer.NewProviderBuilder().
  WithComponents(infer.ComponentF(components.NewWebApp)).
  WithDescription("Deploy a Zerops web application with optional PostgreSQL using a compiled component.").
  WithDisplayName("Zerops App").WithPublisher("Zerops").
  WithHomepage("https://zerops.io").WithRepository("https://github.com/zeropsio/pulumi-zerops").WithLicense("MIT").
  WithKeywords("zerops", "kind/component", "category/cloud").
  WithPluginDownloadURL("github://api.github.com/zeropsio/pulumi-zerops").
  WithLanguageMap(map[string]any{
   "go": map[string]any{"importBasePath": "github.com/zeropsio/pulumi-zerops/sdk-app/go/zeropsapp", "rootPackageName": "zeropsapp", "respectSchemaVersion": true},
   "python": map[string]any{"packageName": "pulumi_zerops_app", "respectSchemaVersion": true, "pyproject": map[string]any{"enabled": true}},
   "nodejs": map[string]any{"packageName": "@zerops/pulumi-app", "respectSchemaVersion": true},
   "csharp": map[string]any{"rootNamespace": "Zerops.Pulumi", "namespaces": map[string]string{"zerops-app": "App"}, "respectSchemaVersion": true},
   "java": map[string]any{"basePackage": "io.zerops", "packages": map[string]string{"zerops-app": "zeropsapp"}, "buildFiles": "gradle", "gradleNexusPublishPluginVersion": "2.0.0"},
  }).Build()
 if err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
 p = withPrimitiveDependency(p)
 if len(os.Args) == 2 && os.Args[1] == "--schema" {
  schema, err := provider.GetSchema(context.Background(), "zerops-app", version, p)
  if err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
  data, err := json.MarshalIndent(schema, "", "  ")
  if err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
  fmt.Println(string(data))
  return
 }
 if err := p.Run(context.Background(), "zerops-app", version); err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
}

// Infer models foreign resource references but does not expose package dependency
// metadata. Attach the descriptor to every schema response, including SDK generation.
func withPrimitiveDependency(p provider.Provider) provider.Provider {
 getSchema := p.GetSchema
 p.GetSchema = func(ctx context.Context, req provider.GetSchemaRequest) (provider.GetSchemaResponse, error) {
  response, err := getSchema(ctx, req)
  if err != nil { return response, err }
  var spec schema.PackageSpec
  if err := json.Unmarshal([]byte(response.Schema), &spec); err != nil { return provider.GetSchemaResponse{}, err }
  spec.Dependencies = []schema.PackageDescriptor{{
   Name: "zerops",
   Version: new(semver.MustParse("0.1.1")),
   DownloadURL: "github://api.github.com/zeropsio/pulumi-zerops",
  }}
  data, err := json.Marshal(spec)
  if err != nil { return provider.GetSchemaResponse{}, err }
  response.Schema = string(data)
  return response, nil
 }
 return p
}
