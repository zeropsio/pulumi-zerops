---
title: Zerops
meta_desc: Manage Zerops projects, services, deployments, storage and networking with Pulumi in your preferred language.
layout: package
---

The `zerops` package manages Zerops projects, services, deployments, storage and networking using Python, Go, C#, Java, TypeScript or YAML. It is built from the Zerops Terraform provider, but does not require Terraform. For an application-first experience with a project, deployment, optional PostgreSQL and a public URL, start with the separate [zerops-app WebApp component](app/_index.md).

## Installation

Install the [Pulumi CLI](https://www.pulumi.com/docs/install/) and the SDK for your program language. YAML needs only the CLI and compiled provider plugin. See [installation and configuration](installation-configuration.md) for pinned commands, plugin downloads and runtime requirements.

{{< chooser language "typescript,python,go,csharp,java,yaml" >}}
{{% choosable language typescript %}}
```sh
npm install @zerops/pulumi@0.1.1
```
{{% /choosable %}}
{{% choosable language python %}}
```sh
python -m pip install pulumi-zerops==0.1.1
```
{{% /choosable %}}
{{% choosable language go %}}
```sh
go get github.com/zeropsio/pulumi-zerops/sdk/go/zerops@v0.1.1
```
{{% /choosable %}}
{{% choosable language csharp %}}
```sh
dotnet add package Zerops.Pulumi --version 0.1.1
```
{{% /choosable %}}
{{% choosable language java %}}
```xml
<dependency>
  <groupId>io.zerops</groupId>
  <artifactId>pulumi</artifactId>
  <version>0.1.1</version>
</dependency>
```
{{% /choosable %}}
{{% choosable language yaml %}}
```sh
pulumi plugin install resource zerops 0.1.1 --server github://api.github.com/zeropsio/pulumi-zerops
```
{{% /choosable %}}
{{< /chooser >}}

## Example Usage

These minimal examples create an explicitly named `LIGHT` project and a service without deploying code. They do not allocate dedicated IPv4. For code deployment and a browser URL, use a [complete quickstart](quickstart-python.md) or the [primitive examples](../examples/).

{{< chooser language "typescript,python,go,csharp,java,yaml" >}}
{{% choosable language typescript %}}
```typescript
import * as zerops from "@zerops/pulumi";

const project = new zerops.Project("project", { name: "hello", mode: "LIGHT" });
const service = new zerops.Service("app", {
    projectId: project.id, hostname: "app", type: "nodejs@22", startWithoutCode: true,
});
export const serviceId = service.id;
```
{{% /choosable %}}
{{% choosable language python %}}
```python
import pulumi
import pulumi_zerops as zerops

project = zerops.Project("project", name="hello", mode="LIGHT")
service = zerops.Service("app", project_id=project.id, hostname="app",
                         type="nodejs@22", start_without_code=True)
pulumi.export("serviceId", service.id)
```
{{% /choosable %}}
{{% choosable language go %}}
```go
package main

import (
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
    "github.com/zeropsio/pulumi-zerops/sdk/go/zerops"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        project, err := zerops.NewProject(ctx, "project", &zerops.ProjectArgs{
            Name: pulumi.String("hello"), Mode: pulumi.String("LIGHT"),
        })
        if err != nil { return err }
        service, err := zerops.NewService(ctx, "app", &zerops.ServiceArgs{
            ProjectId: project.ID(), Hostname: pulumi.String("app"),
            Type: pulumi.String("nodejs@22"), StartWithoutCode: pulumi.Bool(true),
        })
        if err != nil { return err }
        ctx.Export("serviceId", service.ID())
        return nil
    })
}
```
{{% /choosable %}}
{{% choosable language csharp %}}
```csharp
using Pulumi;
using Z = Zerops.Pulumi;

return await Deployment.RunAsync(() => {
    var project = new Z.Project("project", new Z.ProjectArgs {
        Name = "hello", Mode = "LIGHT",
    });
    var service = new Z.Service("app", new Z.ServiceArgs {
        ProjectId = project.Id, Hostname = "app", Type = "nodejs@22", StartWithoutCode = true,
    });
    return new System.Collections.Generic.Dictionary<string, object?> {
        ["serviceId"] = service.Id,
    };
});
```
{{% /choosable %}}
{{% choosable language java %}}
```java
import com.pulumi.Pulumi;
import io.zerops.zerops.Project;
import io.zerops.zerops.ProjectArgs;
import io.zerops.zerops.Service;
import io.zerops.zerops.ServiceArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var project = new Project("project", ProjectArgs.builder()
                .name("hello").mode("LIGHT").build());
            var service = new Service("app", ServiceArgs.builder()
                .projectId(project.id()).hostname("app").type("nodejs@22")
                .startWithoutCode(true).build());
            ctx.export("serviceId", service.id());
        });
    }
}
```
{{% /choosable %}}
{{% choosable language yaml %}}
```yaml
name: hello
runtime: yaml
resources:
  project:
    type: zerops:index:Project
    options:
      version: 0.1.1
    properties:
      name: hello
      mode: LIGHT
  service:
    type: zerops:index:Service
    options:
      version: 0.1.1
    properties:
      projectId: ${project.id}
      hostname: app
      type: nodejs@22
      startWithoutCode: true
outputs:
  serviceId: ${service.id}
```
{{% /choosable %}}
{{< /chooser >}}

## Configuration

Choose either `pulumi login` (Pulumi Cloud) or `pulumi login --local` (local state and a secrets passphrase to retain). Your backend login is separate from Zerops authentication. Set your token securely:

```sh
pulumi config set --secret zerops:token
```

The provider supports `zerops:token`, `zerops:endpoint`, `zerops:organizationId`, `zerops:postgresqlMinDiskGb`, and `zerops:routingForceSync`. The organization ID is membership `clientId`, not membership ID. See [configuration defaults, environment fallbacks and backend choices](installation-configuration.md).

The RAM floors are environment-only settings, not additional Pulumi config keys: `ZEROPS_CLICKHOUSE_MIN_RAM_GB` defaults to `0.375` GB and `ZEROPS_ELASTICSEARCH_MIN_RAM_GB` to `2` GB. Overrides must be finite positive numbers with at most three decimal places. Preview defaults omitted RAM to the floor and rejects smaller configured values; `verticalAutoscaling.minRamGb` and `maxRamGb` are optional/computed.

Dual-mode managed types require explicit `:single@` or `:ha@` modes in services and project-import YAML. For example, use `postgresql:single@16` or `postgresql:ha@16`. When upgrading existing shorthand state, set `type` to the qualified value in `base` for a state-only change without replacement; changing the mode replaces the service.

Secrets are encrypted in state, not excluded from it. Directory deployment uploads every file except `.git` directories: curate sources, and never include secrets or dependency directories. Replacing `ServiceDeploy` starts a new deployment; destroying it does not roll back code. See the [resource and data-function map](../README.md#primitive-resource-map).
