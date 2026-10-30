# Deploy an app with Java

Use Java to describe your infrastructure and deploy a small Python web app to Zerops. The application runs and builds on Zerops: you do not need Python installed locally unless Python is also your chosen infrastructure language.

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/install/) 3.x (tested with 3.273.0).
- JDK 17+ and Maven. Examples target Java 17. The generated SDKs build with JDK 21 and emit Java 11-compatible bytecode. No Node.js installation is needed.
- A Zerops account/organization and an API token from **Settings → Access tokens**, authorized for that organization. No Terraform installation is required.

## 1. Choose a state backend

Choose one before creating the project. This backend identity stores Pulumi state; it does **not** authenticate you to Zerops.

| Pulumi Cloud | Local state, no Pulumi account |
|---|---|
| Run `pulumi login` and follow the browser login. | Run `pulumi login --local`. |
| Pulumi Cloud manages state and its secrets keys. | You own backup/sharing of local state. |
| Requires a Pulumi account; a free tier is available. | Stack creation asks for a secrets passphrase: retain it securely to decrypt state/config later. Automation can supply `PULUMI_CONFIG_PASSPHRASE` securely. |

Secrets are encrypted **in state**, not omitted from it. Protect state backups and your passphrase/keys. More: [installation and configuration](installation-configuration.md).

## 2. Create your project

In an empty directory, run:

```sh
pulumi new https://github.com/zeropsio/pulumi-zerops/tree/v0.1.1/templates/zerops-java
```

Choose a project and stack name when prompted. The Zerops-specific prompts ask for a **secret token** and an explicit **appName** (the Zerops project name). The starter uses `io.zerops:pulumi-app` and defaults to no database. The infrastructure program is `src/main/java/io/zerops/examples/App.java`; the application source is in `app/`.

Set or replace the token securely for the selected stack:

```sh
pulumi config set --secret zerops:token
```

This prompts for the value rather than putting a literal token in your shell history. If the token accesses multiple organizations, set `zerops:organizationId` to your intended organization's membership **clientId**, not membership ID. See [all provider configuration keys](installation-configuration.md#provider-configuration).

## 3. Deploy and open the app

```sh
pulumi up
pulumi stack output url
```

Review the preview, confirm, and open the printed URL in your browser. The app should display its greeting. Build and provisioning can take time; inspect the build log in the Zerops GUI if the deployment fails.

### What gets created and what costs money

`WebApp` creates a `LIGHT` project, one application service/container running `python@3.12`, a deployment from `./app`, and public subdomain access. It does **not** allocate dedicated IPv4. Resources cost money while running; this is a small example, not a free-tier or production availability promise. Review [Zerops pricing](https://zerops.io/pricing).

## 4. Enable PostgreSQL

```sh
pulumi config set database true
pulumi up
```

The added database is **`postgresql:single@16`**, with a **2 GB minimum disk**. It is single-node, not HA, and adds cost. For HA, change the component's `databaseType` input to `postgresql:ha@16` in your program and review the different costs; `databaseType` is not a template config key.

The component injects `DATABASE_URL`, `DB_HOST`, `DB_USER`, `DB_PASSWORD`, and `DB_NAME` as server-side references, plus literal `DB_PORT=5432`. It never reads resolved credentials back. Refresh the browser page: the supplied app reports that `DATABASE_URL` is set, without displaying its value. See [WebApp inputs and outputs](app/_index.md).

## 5. Change your app and redeploy

Edit the greeting in **`app/server.py`**, save, then run:

```sh
pulumi up
pulumi stack output url
```

Open or refresh the URL to see the new greeting. A source change replaces the `ServiceDeploy` child: replacement means **a new deployment**, not deleting the application service. Destroying only that deployment resource does not roll back code.

`source: ./app` resolves relative to the Pulumi program directory. It uploads the directory as-is except every `.git` directory; `.gitignore` is not an upload filter. Curate it: never add secrets, `.env`, `node_modules`, `.venv`, or build caches. If you change the service hostname, also change the setup name in `app/zerops.yml`.

## 6. Clean up

```sh
pulumi destroy
```

Review the deletion preview and confirm. This removes the managed project, services, public access and optional database, including its data. Back up any data you need **before** destroying. Keep your state available until cleanup succeeds.

## Troubleshooting

- **Missing token:** select the right stack, then run `pulumi config set --secret zerops:token`. Alternatively supply `ZEROPS_TOKEN` securely. A Pulumi backend login is not a Zerops login.
- **Organization selection fails:** set `zerops:organizationId` to the intended organization's membership `clientId`, not membership ID; ensure the token has access. It is auto-discovered only when exactly one organization is available.
- **Invalid hostname:** use 1–40 lowercase letters or digits. `db` is reserved when the database is enabled. The hostname must match the `zerops.yml` setup name (default `app`).
- **Asynchronous build/deploy failure:** open the service in the Zerops GUI and inspect its **build log**, correct the application or `app/zerops.yml`, then run `pulumi up` again. The GUI log contains the underlying build error.
- **SDK or plugin cannot be resolved:** check the exact 0.1.1 package/version and [plugin download commands](installation-configuration.md#provider-plugins). Go uses module tags `sdk/v0.1.1` and `sdk-app/v0.1.1`; consumer `go get` commands use `@v0.1.1`, without monorepo replacements.

Continue with [the component reference](app/_index.md), [primitive examples](../examples/), or the [project overview](../README.md).
