package main

import (
 "context"
 "encoding/json"
 "errors"
 "testing"

 provider "github.com/pulumi/pulumi-go-provider"
 "github.com/pulumi/pulumi/pkg/v3/codegen/schema"
)

func TestPrimitiveDependencySchema(t *testing.T) {
 p := withPrimitiveDependency(provider.Provider{GetSchema: func(context.Context, provider.GetSchemaRequest) (provider.GetSchemaResponse, error) {
  return provider.GetSchemaResponse{Schema: `{"name":"zerops-app","version":"0.1.0","resources":{"zerops-app:index:WebApp":{"isComponent":true}}}`}, nil
 }})
 response, err := p.GetSchema(context.Background(), provider.GetSchemaRequest{})
 if err != nil { t.Fatal(err) }
 var spec schema.PackageSpec
 if err := json.Unmarshal([]byte(response.Schema), &spec); err != nil { t.Fatal(err) }
 if spec.Name != "zerops-app" || !spec.Resources["zerops-app:index:WebApp"].IsComponent { t.Fatal("schema wrapper lost package contents") }
 if len(spec.Dependencies) != 1 { t.Fatalf("expected one primitive dependency, got %d",len(spec.Dependencies)) }
 dep := spec.Dependencies[0]
 if dep.Name != "zerops" || dep.Version == nil || dep.Version.String() != "0.1.0" || dep.DownloadURL != "github://api.github.com/zeropsio/pulumi-zerops" { t.Fatalf("incorrect primitive descriptor: %#v", dep) }
}

func TestPrimitiveDependencySchemaErrors(t *testing.T) {
 expected := errors.New("schema unavailable")
 p := withPrimitiveDependency(provider.Provider{GetSchema: func(context.Context, provider.GetSchemaRequest) (provider.GetSchemaResponse,error) { return provider.GetSchemaResponse{},expected }})
 if _,err := p.GetSchema(context.Background(),provider.GetSchemaRequest{}); !errors.Is(err,expected) { t.Fatalf("schema error was lost: %v",err) }
}
