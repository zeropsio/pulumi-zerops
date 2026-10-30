package components

import (
 "os"
 "path/filepath"
 "strings"
 "testing"
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
