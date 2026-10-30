package components

import (
 "os"
 "strings"
 "sync"
 "testing"

 "github.com/pulumi/pulumi/sdk/v3/go/common/resource"
 "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
 "github.com/zeropsio/pulumi-zerops/sdk/go/zerops"
)

type compositionMocks struct { mu sync.Mutex; resources []pulumi.MockResourceArgs }
func (m *compositionMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap,error) { return args.Args,nil }
func (m *compositionMocks) NewResource(args pulumi.MockResourceArgs) (string,resource.PropertyMap,error) {
 m.mu.Lock(); defer m.mu.Unlock(); m.resources = append(m.resources,args)
 outputs := args.Inputs.Copy()
 if strings.Contains(args.TypeToken,"serviceSubdomainAccess") { outputs["urls"] = resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("https://example.com")}) }
 return args.Name+"-id",outputs,nil
}

func TestCompositionAndProviderInheritance(t *testing.T) {
 mocks := &compositionMocks{}
 err := pulumi.RunErr(func(ctx *pulumi.Context) error {
  provider, err := zerops.NewProvider(ctx,"explicit", &zerops.ProviderArgs{Token:pulumi.String("test-token")})
  if err != nil { return err }
  _,err = NewWebApp(ctx,"web",WebAppArgs{AppName:"hello",Runtime:"python@3.12",Database:true,Env:map[string]pulumi.StringInput{"EXTRA":pulumi.ToSecret(pulumi.String("private")).(pulumi.StringOutput)}}, pulumi.Providers(provider))
  return err
 },pulumi.WithMocks("project","stack",mocks))
 if err != nil { t.Fatal(err) }
 count := 0
 for _, r := range mocks.resources {
  if !strings.HasPrefix(r.TypeToken,"zerops:") { continue }
  count++
  if !strings.Contains(r.Provider,"explicit") { t.Errorf("%s lost explicit provider: %s",r.Name,r.Provider) }
  if r.RegisterRPC == nil || !strings.Contains(r.RegisterRPC.Parent,"WebApp::web") { t.Errorf("%s is not parented to component",r.Name) }
  if r.Name == "web-env-EXTRA" && !r.Inputs["content"].IsSecret() { t.Error("extra env lost secrecy") }
  if strings.Contains(r.TypeToken,"serviceDeploy") || strings.Contains(r.TypeToken,"serviceSubdomainAccess") { t.Errorf("created %s without source",r.TypeToken) }
 }
 if count != 10 { t.Errorf("expected project, two services and seven envs, got %d",count) }
}

func TestDeployWaitsForEnvironment(t *testing.T) {
 dir := t.TempDir()
 t.Chdir(dir)
 if err := os.WriteFile("zerops.yml", []byte("zerops:\n  - setup: app\n"), 0600); err != nil { t.Fatal(err) }
 mocks := &compositionMocks{}
 err := pulumi.RunErr(func(ctx *pulumi.Context) error {
  _, err := NewWebApp(ctx, "web", WebAppArgs{AppName:"hello", Runtime:"python@3.12", Source:new("."), Database:true, Env:map[string]pulumi.StringInput{"EXTRA":pulumi.String("value")}})
  return err
 }, pulumi.WithMocks("project","stack",mocks))
 if err != nil { t.Fatal(err) }
 found := false
 for _, r := range mocks.resources {
  if !strings.Contains(r.TypeToken,"serviceDeploy") { continue }
  found = true
  if r.Inputs["sourceDir"].StringValue() != "." { t.Error("relative source path was not preserved") }
  count := 0
  for _, dep := range r.RegisterRPC.Dependencies { if strings.Contains(dep,"ServiceEnv::web-env-") { count++ } }
  if count != 7 { t.Errorf("deployment waits on %d env resources; wanted 7",count) }
 }
 if !found { t.Fatal("deployment missing") }
}
