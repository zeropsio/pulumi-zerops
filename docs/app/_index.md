---
title: Zerops App
meta_desc: Deploy a Zerops web application and optional PostgreSQL database with the multi-language WebApp component.
layout: package
---

The `zerops-app` package is a **kind/component** package: its `WebApp` component composes resources from the `zerops` provider into an application environment. It uses a compiled plugin, so you only need your chosen Pulumi program language locally, not Go or a separate JavaScript runtime.

## Installation

Install the component SDK for [Python](../quickstart-python.md), [Go](../quickstart-go.md), [C#](../quickstart-csharp.md), [Java](../quickstart-java.md), or [TypeScript](../quickstart-typescript.md). [YAML](../quickstart-yaml.md) needs no language SDK. See [installation and configuration](../installation-configuration.md) for package identities and plugin installation; the component SDK depends on the primitive SDK.

## Example Usage

```yaml
name: hello
runtime: yaml
resources:
  web:
    type: zerops-app:index:WebApp
    options:
      version: 0.1.0
    properties:
      appName: hello
      runtime: python@3.12
      source: ./app
      database: true
outputs:
  url: ${web.url}
  projectId: ${web.projectId}
  serviceId: ${web.serviceId}
```

The `app/` directory must contain the application and a `zerops.yml` or `zerops.yaml` with setup `app` (or supply `zeropsYaml`). The [templates](../../templates/) include a complete working application.

## Inputs

Names below are schema/YAML names; SDKs use their language's naming conventions.

| Input | Meaning / default |
|---|---|
| `appName` | Required explicit Zerops project name; no automatic naming |
| `runtime` | Required service type, such as `python@3.12` or `nodejs@22` |
| `source` | Optional curated application directory; relative to the Pulumi program directory |
| `zeropsYaml` | Optional deployment YAML content; otherwise read from the source directory |
| `hostname` | `app`; 1–40 lowercase letters or digits; `db` is reserved when database is enabled |
| `database` | `false`; enable PostgreSQL |
| `databaseType` | `postgresql:single@16`; use explicit `postgresql:ha@16` for HA |
| `env` | Additional runtime environment values; secret inputs remain secret; database connection names are reserved when enabled |
| `publicAccess` | `true`; create public subdomain access after deploying code |

Without `source`, the component creates the service without code and does not create a deployment or public access, even if `publicAccess` is true. Match your deployment YAML setup name to `hostname`.

## Outputs

| Output | Meaning |
|---|---|
| `url` | HTTPS public URL, or empty when no source/public access is configured |
| `projectId` | Created project ID |
| `serviceId` | Application service ID |
| `databaseServiceId` | Database service ID; empty when disabled |
| `connectionEnvNames` | Names of injected database connection variables, not credential values |
| `project` | Typed primitive `Project` child resource |
| `service` | Typed primitive `Service` child resource |

## Database connections, costs and deployment

The component creates a `LIGHT` project and one application container, with no dedicated IPv4. The optional default PostgreSQL service is **single-node**, with a provider minimum disk of **2 GB**. This is not an HA configuration or a free-tier promise; review [pricing](https://zerops.io/pricing) before deploying. Choose `postgresql:ha@16` explicitly when you need HA.

With the database enabled, these application variables are created:

| Variable | Server-side reference / literal |
|---|---|
| `DATABASE_URL` | `${db_connectionString}` |
| `DB_HOST` | `${db_hostname}` |
| `DB_PORT` | `5432` (literal) |
| `DB_USER` | `${db_user}` |
| `DB_PASSWORD` | `${db_password}` |
| `DB_NAME` | `${db_dbName}` |

Zerops resolves the references server-side; the component never reads the resolved credentials back. Secret inputs still end up encrypted in Pulumi state, not excluded from it.

The source directory is uploaded as-is except every `.git` directory. Curate it: exclude secrets, `.env`, `node_modules`, `.venv`, and caches yourself; `.gitignore` does not filter uploads. A source change replaces the `ServiceDeploy` child and triggers a **new deploy**; deleting that deployment resource does not roll back code. `pulumi destroy` removes the managed service/project and database data. Review data-loss implications first.

## Configuration

The component's children inherit the `zerops` provider, including `zerops:token` and `zerops:organizationId`. You can also supply an explicit `zerops` provider through the component's providers option. See [all five provider configuration keys](../installation-configuration.md#provider-configuration).

If a deployment fails asynchronously, inspect the service's build log in the Zerops GUI. See the language quickstarts for token, organization and hostname troubleshooting.
