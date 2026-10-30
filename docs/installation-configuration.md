# Installation and configuration

Install the [Pulumi CLI](https://www.pulumi.com/docs/install/) (3.x; tested with 3.273.0), then use your preferred infrastructure language. Terraform and Node.js are not prerequisites for Python, Go, .NET, Java or YAML programs.

## Choose your state backend

There are **two separate identities**: your Pulumi backend stores stack state; your Zerops token authorizes infrastructure changes. Logging in to Pulumi does not log in to Zerops.

| Pulumi Cloud | Local backend (no Pulumi account) |
|---|---|
| Run `pulumi login` and follow the browser/account prompt. | Run `pulumi login --local`. |
| State and secrets-key management are handled by Pulumi Cloud. | State stays on your computer. You must back it up and arrange sharing yourself. |
| Uses a Pulumi account, including the free tier. | Stack creation prompts for a secrets passphrase. Keep it safe: you need it to decrypt state/config. For automation, supply `PULUMI_CONFIG_PASSPHRASE` securely. |

Secret values are encrypted **in state**, not excluded from state. Protect state, backups, and secrets keys. See [Pulumi backends](https://www.pulumi.com/docs/iac/concepts/state-and-backends/).

## Install a language SDK

Start with the `zerops-app` component SDK for an application; it depends on `zerops` primitives. Install the primitive SDK directly if you only need primitives. These are alternatives, not an instruction to install both manually.

### Python

Requires Python >= 3.10. In your project's virtual environment:

```sh
python -m pip install pulumi-zerops-app==0.1.0
# Primitives only:
python -m pip install pulumi-zerops==0.1.0
```

Imports: `pulumi_zerops_app` and `pulumi_zerops`.

### Go

Requires Go >= 1.26.6 with the current modules:

```sh
go get github.com/zeropsio/pulumi-zerops/sdk-app/go/zeropsapp@v0.1.0
# Primitives only:
go get github.com/zeropsio/pulumi-zerops/sdk/go/zerops@v0.1.0
```

The modules are `github.com/zeropsio/pulumi-zerops/sdk-app` and `github.com/zeropsio/pulumi-zerops/sdk`, released with tags `sdk-app/v0.1.0` and `sdk/v0.1.0`. `components/` and `sdk-app/` retain monorepo `replace` directives for development; Go does not apply dependency modules' replacements to your project. Consumer programs do not need these replacements.

### C# / .NET

Use the .NET 8 or 10 SDK. Both language SDKs target `net6.0`; the starter targets .NET 8.

```sh
dotnet add package Zerops.Pulumi.App --version 0.1.0
# Primitives only:
dotnet add package Zerops.Pulumi --version 0.1.0
```

Namespaces match the package names: `Zerops.Pulumi.App` and `Zerops.Pulumi`.

### Java

Use Java 17+ and Maven for the starter. The SDKs themselves build using JDK 21 and produce Java 11-compatible bytecode. Add the appropriate dependency to `pom.xml`:

```xml
<dependency>
  <groupId>io.zerops</groupId>
  <artifactId>pulumi-app</artifactId>
  <version>0.1.0</version>
</dependency>
```

For primitives, use artifact `pulumi` instead of `pulumi-app`. Resolve dependencies with:

```sh
mvn dependency:resolve
```

Java packages are `io.zerops.zeropsapp` and `io.zerops.zerops` respectively, not their Maven artifact names.

### TypeScript / Node.js

Node.js 22 LTS is tested:

```sh
npm install @zerops/pulumi-app@0.1.0
# Primitives only:
npm install @zerops/pulumi@0.1.0
```

### YAML

No language runtime or language SDK is needed. Use `zerops-app:index:WebApp` for the component and short primitive tokens such as `zerops:index:Project`.

## Provider plugins

Language SDKs carry plugin download metadata. The CLI downloads compiled resource plugins from GitHub releases; you do not need Go to run them. For a pinned YAML program, set resource `options.version: 0.1.0`. To install explicitly (or prepare a plugin cache):

```sh
pulumi plugin install resource zerops 0.1.0 --server github://api.github.com/zeropsio/pulumi-zerops
pulumi plugin install resource zerops-app 0.1.0 --server github://api.github.com/zeropsio/pulumi-zerops
```

`WebApp` uses both plugins. A primitives-only program only needs `zerops`.

## Authenticate to Zerops

Create a Zerops API token in **Settings → Access tokens** with access to the organization you intend to manage. Within an initialized stack:

```sh
pulumi config set --secret zerops:token
```

The command prompts without placing a literal token in shell history. Alternatively, supply `ZEROPS_TOKEN` from a secure environment. Do not commit plaintext tokens or source directories containing credentials.

## Provider configuration

All five keys belong to **`zerops`**, including when you use `zerops-app`. Explicit values take precedence over environment fallbacks, which take precedence over defaults.

| Pulumi config key | Environment fallback | Default and meaning | Secret |
|---|---|---|---|
| `zerops:token` | `ZEROPS_TOKEN` | Required after resolution; API integration/access token; no sudo escalation | Yes |
| `zerops:endpoint` | `ZEROPS_ENDPOINT` | `https://api.app-prg1.zerops.io`; must be an HTTPS origin | No |
| `zerops:organizationId` | `ZEROPS_ORGANIZATION_ID` | Discovered when the token has exactly one organization; otherwise specify explicitly | No |
| `zerops:postgresqlMinDiskGb` | `ZEROPS_POSTGRESQL_MIN_DISK_GB` | `2` GB; PostgreSQL minimum disk floor, at most three decimal places | No |
| `zerops:routingForceSync` | `ZEROPS_ROUTING_FORCE_SYNC` | `false`; whether to synchronize foreign staged routing changes | No |

**Organization ID means membership `clientId`, not the membership ID.** If your token can access multiple organizations, copy the intended organization's client ID and set it with `pulumi config set zerops:organizationId <clientId>`. Do not paste an example or unrelated project/service ID.

The default disk floor avoids provisioning failures for PostgreSQL below 2 GB. The component's default database is `postgresql:single@16`; use `databaseType: postgresql:ha@16` explicitly for HA. Neither option is free, and HA has different costs. Leave `routingForceSync` false unless you intend to take ownership of other staged routing changes.

Component-specific inputs such as `appName`, `source`, `database` and `databaseType` are **not provider configuration keys**. The templates expose the first and the database toggle through their own stack config; see the [component reference](app/_index.md).

## Next steps

Follow a complete quickstart: [Python](quickstart-python.md), [Go](quickstart-go.md), [C#](quickstart-csharp.md), [Java](quickstart-java.md), [TypeScript](quickstart-typescript.md), or [YAML](quickstart-yaml.md). Return to the [project overview](../README.md) for deployment semantics and the resource map.
