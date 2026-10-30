# Zerops for Pulumi

Deploy your application to [Zerops](https://zerops.io) using Python, Go, C#, Java, TypeScript or YAML. Keep application infrastructure alongside your code, preview changes, and deploy with `pulumi up`.

Start with **`zerops-app` / `WebApp`**: one component creates a project, an application service, an optional PostgreSQL database, a deployment and a public URL. Use **`zerops`** primitives when you need direct control over services, storage or networking. The primitive provider bridges the [Zerops Terraform provider](https://github.com/zeropsio/terraform-provider-zerops); you do not need Terraform installed.

## Start in your language

[Python](docs/quickstart-python.md) · [Go](docs/quickstart-go.md) · [C#/.NET](docs/quickstart-csharp.md) · [Java](docs/quickstart-java.md) · [TypeScript](docs/quickstart-typescript.md) · [YAML](docs/quickstart-yaml.md)

Each starter deploys the same small Python web application. The language you use to describe infrastructure does not dictate your application's runtime. Only your chosen infrastructure language and the Pulumi CLI are needed locally; Zerops builds and runs the application.

For example, after [choosing a state backend](docs/installation-configuration.md#choose-your-state-backend):

```sh
pulumi new https://github.com/zeropsio/pulumi-zerops/tree/v0.1.1/templates/zerops-python
pulumi config set --secret zerops:token
pulumi up
pulumi stack output url
```

The template also prompts for a secret Zerops token and an explicit `appName`. PostgreSQL is off initially; enable it with `pulumi config set database true`, then run `pulumi up` again. Review costs before deploying; resources run until you remove them. Run `pulumi destroy` when finished (this deletes application and database data).

## Packages and compatibility

Both packages are version **0.1.1**. The component SDK brings in the primitive SDK as a dependency.

| Language | `zerops` primitives | `zerops-app` component |
|---|---|---|
| Python | PyPI `pulumi-zerops`; import `pulumi_zerops` | PyPI `pulumi-zerops-app`; import `pulumi_zerops_app` |
| Go | module `github.com/zeropsio/pulumi-zerops/sdk`; import `github.com/zeropsio/pulumi-zerops/sdk/go/zerops` | module `github.com/zeropsio/pulumi-zerops/sdk-app`; import `github.com/zeropsio/pulumi-zerops/sdk-app/go/zeropsapp` |
| C# | NuGet and namespace `Zerops.Pulumi` | NuGet and namespace `Zerops.Pulumi.App` |
| Java | Maven `io.zerops:pulumi`; package `io.zerops.zerops` | Maven `io.zerops:pulumi-app`; package `io.zerops.zeropsapp` |
| TypeScript / Node.js | npm `@zerops/pulumi` | npm `@zerops/pulumi-app` |
| YAML | plugin `zerops`; e.g. `zerops:index:Project` | plugin `zerops-app`; `zerops-app:index:WebApp` |

Go modules use release tags **`sdk/v0.1.1`** and **`sdk-app/v0.1.1`**; use `@v0.1.1` in `go get`, not the tag's directory prefix. The `replace` directives in `components/go.mod` and `sdk-app/go.mod` are for monorepo development. Go ignores dependency modules' replacements, so consumers never inherit them.

| Pulumi provider | Terraform provider | Terraform bridge | Pulumi CLI |
|---|---|---|---|
| `pulumi-zerops` 0.1.1 | `terraform-provider-zerops` v0.1.1 | v3.142.0 | >= 3.x; tested with 3.273.0 |

Python requires >= 3.10. Go modules currently require Go >= 1.26.6. The .NET SDK targets `net6.0` and works with .NET 8 and 10; templates target .NET 8. Java SDKs compile with JDK 21 and emit Java 11-compatible bytecode; examples and templates target Java 17. Node.js 22 LTS is the tested Node runtime.

## What WebApp creates

A `LIGHT` project with an explicit name, a single-container runtime service, optional PostgreSQL, environment variables, a directory deployment, and public subdomain access after deployment. It does **not** allocate dedicated IPv4. The optional default database is **`postgresql:single@16`**, with a **2 GB minimum disk**. This is a small, non-HA starting point, not a free tier or a production availability promise. For HA, explicitly set `databaseType` to `postgresql:ha@16`; availability and price differ. See [Zerops pricing](https://zerops.io/pricing).

Database connections use server-side references: `DATABASE_URL`, `DB_HOST`, `DB_USER`, `DB_PASSWORD`, and `DB_NAME` reference database service values; `DB_PORT` is the literal `5432`. The component never reads resolved credentials back. It exports connection variable **names**, not passwords. See the [component reference](docs/app/_index.md).

## Deployment and secrets

- `WebApp.source` and primitive `ServiceDeploy.sourceDir` are directory paths, not Pulumi archives. Relative paths resolve from the Pulumi program directory.
- The directory is uploaded **as-is except every `.git` directory**. Curate `app/`: do not include secrets, `.env`, `node_modules`, `.venv`, or build caches. `.gitignore` is not an upload filter.
- Source changes replace `ServiceDeploy`, meaning **a new deployment**. Destroying only that resource does not roll back the app. Destroying its service/project cleans up the environment.
- Secret configuration and secret resource values are encrypted in Pulumi state, not omitted from it. Protect state backups and your backend's encryption keys/passphrase. Do not export secrets as ordinary stack outputs.
- Zerops builds asynchronously. If deployment fails, open the service's build log in the Zerops GUI for the actual build error.

## Primitive resource map

The schema contains **19 resources**. The short YAML forms below correspond to canonical tokens such as `zerops:index/project:Project`.

| Resource | Short token |
|---|---|
| Integration token | `zerops:index:IntegrationToken` |
| Object storage | `zerops:index:ObjectStorage` |
| Project | `zerops:index:Project` |
| Dedicated IPv4 | `zerops:index:ProjectDedicatedIpv4` |
| Project environment variable | `zerops:index:ProjectEnv` |
| Project import | `zerops:index:ProjectImport` |
| L7 balancer configuration | `zerops:index:ProjectL7BalancerConfig` |
| Prometheus configuration | `zerops:index:ProjectPrometheusConfig` |
| Remote logging | `zerops:index:ProjectRemoteLogging` |
| VPN peer | `zerops:index:ProjectVpnPeer` |
| Public HTTP routing | `zerops:index:PublicHttpRouting` |
| Public port routing | `zerops:index:PublicPortRouting` |
| Service | `zerops:index:Service` |
| Backup configuration | `zerops:index:ServiceBackupConfig` |
| Deployment | `zerops:index:ServiceDeploy` |
| Service environment variable | `zerops:index:ServiceEnv` |
| Git integration | `zerops:index:ServiceGitIntegration` |
| Restart | `zerops:index:ServiceRestart` |
| Subdomain access | `zerops:index:ServiceSubdomainAccess` |

There are **9 data functions**, plus the bridge's provider-bound `terraformConfig` method (not a Zerops lookup).

| Lookup | Canonical function token |
|---|---|
| Locations | `zerops:index/getLocations:getLocations` |
| Organization | `zerops:index/getOrganization:getOrganization` |
| Project | `zerops:index/getProject:getProject` |
| Project export | `zerops:index/getProjectExport:getProjectExport` |
| Projects | `zerops:index/getProjects:getProjects` |
| Regions | `zerops:index/getRegions:getRegions` |
| Service | `zerops:index/getService:getService` |
| Service environment variables | `zerops:index/getServiceEnvVars:getServiceEnvVars` |
| Service types | `zerops:index/getServiceTypes:getServiceTypes` |

## More information

- [Installation, authentication and provider configuration](docs/installation-configuration.md)
- [Primitive package overview](docs/_index.md) and [WebApp component overview](docs/app/_index.md)
- [Working primitive examples](examples/) and [application starter templates](templates/)
- [Changelog](CHANGELOG.md), [contributing](CONTRIBUTING.md), and [private security reporting](SECURITY.md)
