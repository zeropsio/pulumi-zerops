package components

import (
 "os"
 "path/filepath"
 "strings"
 "testing"

 "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func TestValidation(t *testing.T) {
 for _, tc := range []struct{name string; args WebAppArgs; message string}{
  {"name", WebAppArgs{Runtime:"python@3.12"}, "appName is required"},
  {"runtime", WebAppArgs{AppName:"hello"}, "runtime is required"},
  {"hostname", WebAppArgs{AppName:"hello",Runtime:"python@3.12",Hostname:new("bad-name")}, "hostname"},
  {"ha shorthand", WebAppArgs{AppName:"hello",Runtime:"python@3.12",Database:true,DatabaseType:new("postgresql@16")}, "explicit mode"},
  {"other database", WebAppArgs{AppName:"hello",Runtime:"python@3.12",Database:true,DatabaseType:new("mysql@8")}, "postgresql:"},
  {"unknown mode", WebAppArgs{AppName:"hello",Runtime:"python@3.12",Database:true,DatabaseType:new("postgresql:other@16")}, "explicit mode"},
  {"missing version", WebAppArgs{AppName:"hello",Runtime:"python@3.12",Database:true,DatabaseType:new("postgresql:single@")}, "explicit mode"},
  {"reserved hostname", WebAppArgs{AppName:"hello",Runtime:"python@3.12",Database:true,Hostname:new("db")}, "reserved"},
 } { t.Run(tc.name, func(t *testing.T) { _,_,_,err := normalize(tc.args); if err == nil || !strings.Contains(err.Error(),tc.message) { t.Fatalf("expected %q, got %v",tc.message,err) } }) }
}

func TestDatabaseModes(t *testing.T) {
 for _, mode := range []string{"single", "ha"} {
  _, _, _, err := normalize(WebAppArgs{AppName:"hello", Runtime:"python@3.12", Database:true, DatabaseType:new("postgresql:"+mode+"@16")})
  if err != nil { t.Fatalf("mode %s: %v", mode, err) }
 }
}

func TestSourceAndDefaults(t *testing.T) {
 a, content, mismatch, err := normalize(WebAppArgs{AppName:"hello",Runtime:"python@3.12"})
 if err != nil || content != "" || mismatch || *a.Hostname != "app" || *a.DatabaseType != "postgresql:single@16" || !*a.PublicAccess { t.Fatalf("incorrect defaults: %#v %v",a,err) }
 dir := t.TempDir()
 a.Source = &dir
 if _,_,_,err := normalize(a); err == nil || !strings.Contains(err.Error(),"no zerops.yml") { t.Fatalf("missing YAML accepted: %v",err) }
 text := "zerops:\n  - setup: app\n"
 if err := os.WriteFile(filepath.Join(dir,"zerops.yaml"), []byte(text),0600); err != nil { t.Fatal(err) }
 _, content, mismatch, err = normalize(a)
 if err != nil || content != text || mismatch { t.Fatalf("file loading: %q %v %v",content,mismatch,err) }
 a.ZeropsYaml = new("zerops:\n  - setup: other\n")
 _,_,mismatch,err = normalize(a)
 if err != nil || !mismatch { t.Fatalf("setup mismatch not detected: %v %v",mismatch,err) }
 a.ZeropsYaml = new("[")
 if _,_,_,err = normalize(a); err == nil { t.Fatal("invalid YAML accepted") }
}

func TestDatabaseReferences(t *testing.T) {
 for key, value := range databaseEnv {
  if key == "DB_PORT" { if value != "5432" { t.Fatal(value) }; continue }
  if !strings.HasPrefix(value,"${db_") || !strings.HasSuffix(value,"}") { t.Fatalf("%s is not a server-side reference",key) }
 }
}

func TestProductionLiveValidation(t *testing.T) {
 t.Setenv("ZEROPS_SANDBOX", "prod")
 for _, tc := range []struct{name, project, hostname, databaseType string; valid bool}{
  {"small CI app", "tfacc-pulumi-test", "app", "postgresql:single@16", true},
  {"foreign project", "customer", "app", "postgresql:single@16", false},
  {"HA database", "tfacc-pulumi-test", "app", "postgresql:ha@16", false},
 } {
  t.Run(tc.name, func(t *testing.T) {
   _, _, _, err := normalize(WebAppArgs{AppName:tc.project, Runtime:"python@3.12", Hostname:&tc.hostname, Database:true, DatabaseType:&tc.databaseType})
   if (err == nil) != tc.valid { t.Fatalf("valid=%v, error=%v", tc.valid, err) }
  })
 }
}

func TestProductionDatabaseResourceArgs(t *testing.T) {
 t.Setenv("ZEROPS_SANDBOX", "prod")
 mocks := &compositionMocks{}
 err := pulumi.RunErr(func(ctx *pulumi.Context) error {
  _, err := NewWebApp(ctx, "web", WebAppArgs{AppName:"tfacc-pulumi-test", Runtime:"python@3.12", Database:true})
  return err
 }, pulumi.WithMocks("project", "stack", mocks))
 if err != nil { t.Fatal(err) }
 foundDB, foundRuntime := false, false
 for _, r := range mocks.resources {
  switch r.Name {
  case "web-database":
   foundDB = true
   if r.Inputs["type"].StringValue() != "postgresql:single@16" { t.Errorf("unexpected database type: %v", r.Inputs["type"]) }
   if horizontal, ok := r.Inputs["horizontalAutoscaling"]; ok && !horizontal.IsNull() { t.Errorf("fixed-topology database has horizontal autoscaling: %v", horizontal) }
   vertical := r.Inputs["verticalAutoscaling"].ObjectValue()
   if vertical["minDiskGb"].NumberValue() != 2 { t.Errorf("database minimum disk must be 2 GB: %v", vertical) }
  case "web-service":
   foundRuntime = true
   horizontal := r.Inputs["horizontalAutoscaling"].ObjectValue()
   if horizontal["minContainers"].NumberValue() != 1 || horizontal["maxContainers"].NumberValue() != 1 { t.Errorf("runtime must retain one container: %v", horizontal) }
  }
 }
 if !foundDB || !foundRuntime { t.Fatalf("missing generated database/runtime resources: database=%v runtime=%v", foundDB, foundRuntime) }
}
