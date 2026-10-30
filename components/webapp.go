package components

import (
 "fmt"
 "os"
 "path/filepath"
 "regexp"
 "sort"
 "strings"

 "github.com/pulumi/pulumi-go-provider/infer"
 "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
 "github.com/zeropsio/pulumi-zerops/sdk/go/zerops"
 "gopkg.in/yaml.v3"
)

type WebAppArgs struct {
 AppName string `pulumi:"appName"`
 Runtime string `pulumi:"runtime"`
 Source *string `pulumi:"source,optional"`
 ZeropsYaml *string `pulumi:"zeropsYaml,optional"`
 Hostname *string `pulumi:"hostname,optional"`
 Database bool `pulumi:"database,optional"`
 DatabaseType *string `pulumi:"databaseType,optional"`
 Env map[string]pulumi.StringInput `pulumi:"env,optional"`
 PublicAccess *bool `pulumi:"publicAccess,optional"`
}

func (a *WebAppArgs) Annotate(an infer.Annotator) {
 an.Describe(&a.AppName, "Explicit Zerops project name (no automatic naming).")
 an.Describe(&a.Runtime, "Runtime service type, for example python@3.12 or nodejs@22.")
 an.Describe(&a.Source, "Curated application directory relative to the Pulumi program. All files except .git directories are uploaded; do not include secrets or dependencies.")
 an.Describe(&a.Env, "Additional runtime environment values; secret inputs remain secret. Database connection names are reserved when database is enabled.")
 an.SetDefault(&a.Hostname, "app")
 an.SetDefault(&a.DatabaseType, "postgresql:single@16")
 an.SetDefault(&a.PublicAccess, true)
}

type WebApp struct {
 pulumi.ResourceState
 Url pulumi.StringOutput `pulumi:"url"`
 ProjectId pulumi.StringOutput `pulumi:"projectId"`
 ServiceId pulumi.StringOutput `pulumi:"serviceId"`
 DatabaseServiceId pulumi.StringOutput `pulumi:"databaseServiceId,optional"`
 ConnectionEnvNames pulumi.StringArrayOutput `pulumi:"connectionEnvNames"`
 Project *zerops.Project `pulumi:"project" provider:"type=zerops@0.1.0:index/project:Project"`
 Service *zerops.Service `pulumi:"service" provider:"type=zerops@0.1.0:index/service:Service"`
}

func (*WebApp) Annotate(a infer.Annotator) { a.SetToken("index", "WebApp") }

var hostnamePattern = regexp.MustCompile(`^[a-z0-9]{1,40}$`)

// PostgreSQL exposes these keys; it does not expose a port key. Its native port is 5432.
var databaseEnv = map[string]string{
 "DATABASE_URL": "${db_connectionString}",
 "DB_HOST": "${db_hostname}",
 "DB_PORT": "5432",
 "DB_USER": "${db_user}",
 "DB_PASSWORD": "${db_password}",
 "DB_NAME": "${db_dbName}",
}

func normalize(a WebAppArgs) (WebAppArgs, string, bool, error) {
 if strings.TrimSpace(a.AppName) == "" { return a, "", false, fmt.Errorf("appName is required: provide an explicit Zerops project name") }
 if strings.TrimSpace(a.Runtime) == "" { return a, "", false, fmt.Errorf("runtime is required: use a runtime service type such as python@3.12") }
 if a.Hostname == nil { a.Hostname = new("app") }
 if !hostnamePattern.MatchString(*a.Hostname) { return a, "", false, fmt.Errorf("hostname must contain 1–40 lowercase letters or digits") }
 if a.DatabaseType == nil { a.DatabaseType = new("postgresql:single@16") }
 if a.Database {
  if !strings.HasPrefix(*a.DatabaseType, "postgresql:") { return a, "", false, fmt.Errorf("databaseType must use an explicit postgresql: family, such as postgresql:single@16; postgresql@16 creates HA, not single-node PostgreSQL") }
  if *a.Hostname == "db" { return a, "", false, fmt.Errorf("hostname db is reserved for the database; choose another application hostname") }
 }
 if a.PublicAccess == nil { a.PublicAccess = new(true) }
 if a.Source == nil { return a, "", false, nil }
 if strings.TrimSpace(*a.Source) == "" { return a, "", false, fmt.Errorf("source must name an application directory, or omit it to create a service without code") }
 // Both resource plugins inherit the engine's program working directory. Keep relative
 // paths relative so state is portable between checkouts and remote runners.
 source := filepath.Clean(*a.Source)
 info, err := os.Stat(source)
 if err != nil || !info.IsDir() { return a, "", false, fmt.Errorf("source %q must be an existing application directory", source) }
 a.Source = &source
 var content string
 if a.ZeropsYaml != nil { content = *a.ZeropsYaml } else {
  for _, name := range []string{"zerops.yml", "zerops.yaml"} {
   data, err := os.ReadFile(filepath.Join(source, name))
   if err == nil { content = string(data); break }
   if !os.IsNotExist(err) { return a, "", false, fmt.Errorf("reading deployment YAML: %w", err) }
  }
  if content == "" { return a, "", false, fmt.Errorf("source %q has no zerops.yml or zerops.yaml; add one or supply zeropsYaml", source) }
 }
 var config struct { Zerops []struct { Setup string `yaml:"setup"` } `yaml:"zerops"` }
 if err := yaml.Unmarshal([]byte(content), &config); err != nil { return a, "", false, fmt.Errorf("invalid zeropsYaml: %w", err) }
 mismatch := true
 for _, setup := range config.Zerops { if setup.Setup == *a.Hostname { mismatch = false } }
 return a, content, mismatch, nil
}

func NewWebApp(ctx *pulumi.Context, name string, args WebAppArgs, opts ...pulumi.ResourceOption) (*WebApp, error) {
 a, deployYaml, mismatch, err := normalize(args)
 if err != nil { return nil, err }
 if mismatch { if err := ctx.Log.Warn(fmt.Sprintf("zerops.yml has no setup named %q; deployment selects the hostname, so update the setup name or hostname", *a.Hostname), nil); err != nil { return nil, err } }
 w := &WebApp{}
 if err := ctx.RegisterComponentResource("zerops-app:index:WebApp", name, w, opts...); err != nil { return nil, err }
 // Parent inheritance preserves explicit providers["zerops"] and otherwise uses
 // the program's default zerops provider, including zerops:token configuration.
 child := pulumi.Parent(w)
 project, err := zerops.NewProject(ctx, name+"-project", &zerops.ProjectArgs{Name: pulumi.String(a.AppName), Mode: pulumi.String("LIGHT")}, child)
 if err != nil { return nil, err }
 w.Project, w.ProjectId = project, project.ID().ToStringOutput()
 var db *zerops.Service
 if a.Database {
  db, err = zerops.NewService(ctx, name+"-database", &zerops.ServiceArgs{ProjectId: project.ID(), Hostname: pulumi.String("db"), Type: pulumi.String(*a.DatabaseType)}, child)
  if err != nil { return nil, err }
  w.DatabaseServiceId = db.ID().ToStringOutput()
 } else { w.DatabaseServiceId = pulumi.String("").ToStringOutput() }
 service, err := zerops.NewService(ctx, name+"-service", &zerops.ServiceArgs{
  ProjectId: project.ID(), Hostname: pulumi.String(*a.Hostname), Type: pulumi.String(a.Runtime), StartWithoutCode: pulumi.Bool(true),
  HorizontalAutoscaling: &zerops.ServiceHorizontalAutoscalingArgs{MinContainers: pulumi.Int(1), MaxContainers: pulumi.Int(1)},
 }, child)
 if err != nil { return nil, err }
 w.Service, w.ServiceId = service, service.ID().ToStringOutput()
 var envs []pulumi.Resource
 names := []string{}
 if db != nil {
  keys := make([]string, 0, len(databaseEnv)); for key := range databaseEnv { keys = append(keys, key) }; sort.Strings(keys)
  for _, key := range keys {
   env, err := zerops.NewServiceEnv(ctx, name+"-env-"+key, &zerops.ServiceEnvArgs{ServiceId: service.ID(), Key: pulumi.String(key), Content: pulumi.String(databaseEnv[key])}, child, pulumi.DependsOn([]pulumi.Resource{db}))
   if err != nil { return nil, err }; envs = append(envs, env); names = append(names, key)
  }
 }
 keys := make([]string, 0, len(a.Env))
 for key := range a.Env { keys = append(keys, key) }
 sort.Strings(keys)
 for _, key := range keys {
  if a.Database { if _, reserved := databaseEnv[key]; reserved { return nil, fmt.Errorf("env.%s is reserved for the database connection", key) } }
  env, err := zerops.NewServiceEnv(ctx, name+"-env-"+key, &zerops.ServiceEnvArgs{ServiceId: service.ID(), Key: pulumi.String(key), Content: a.Env[key]}, child)
  if err != nil { return nil, err }
  envs = append(envs, env)
 }
 return finish(ctx, name, w, a, pulumi.String(deployYaml), envs, child, names)
}

func finish(ctx *pulumi.Context, name string, w *WebApp, a WebAppArgs, deployYaml pulumi.StringInput, envs []pulumi.Resource, child pulumi.ResourceOption, names []string) (*WebApp, error) {
 w.ConnectionEnvNames = pulumi.ToStringArray(names).ToStringArrayOutput()
 w.Url = pulumi.String("").ToStringOutput()
 if a.Source == nil { return w, nil }
 deploy, err := zerops.NewServiceDeploy(ctx, name+"-deploy", &zerops.ServiceDeployArgs{ServiceId: w.Service.ID(), SourceDir: pulumi.String(*a.Source), ZeropsYaml: deployYaml, Setup: pulumi.String(*a.Hostname)}, child, pulumi.DependsOn(envs))
 if err != nil { return nil, err }
 if *a.PublicAccess {
  access, err := zerops.NewServiceSubdomainAccess(ctx, name+"-access", &zerops.ServiceSubdomainAccessArgs{ServiceId: w.Service.ID()}, child, pulumi.DependsOn([]pulumi.Resource{deploy}))
  if err != nil { return nil, err }
  w.Url = access.Urls.ApplyT(func(urls []string) string { for _, url := range urls { if strings.HasPrefix(url, "https://") { return url } }; return "" }).(pulumi.StringOutput)
 }
 return w, nil
}
