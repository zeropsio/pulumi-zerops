# Contributing

Report ordinary bugs and feature requests through [GitHub issues](https://github.com/zeropsio/pulumi-zerops/issues). Include package/plugin versions, your language, a minimal program and sanitized diagnostics. Use [private reporting](SECURITY.md) for security issues; never attach tokens or raw state.

## Repository structure

- `provider/`: Terraform Plugin Framework bridge, token/language metadata and primitive schema.
- `components/`: Go implementation of the compiled `zerops-app` / `WebApp` component and its schema.
- `sdk/` and `sdk-app/`: committed generated SDKs for both packages.
- `templates/`: published-identity application starters; `examples/`: runnable programs and credential-gated acceptance tests.
- `docs/`: user-facing installation, Registry overview and quickstarts.

Resource CRUD, validation and shared planning behavior belong in [terraform-provider-zerops](https://github.com/zeropsio/terraform-provider-zerops). Keep bridge-only naming and SDK metadata in `provider/resources.go`; keep app composition in `components/`. Do not fork shared resource rules in a second implementation.

## Toolchain

Use the Go version required by the module `go.mod` files (currently >= 1.26.6), Pulumi CLI 3.273.0, Python >= 3.10 with `uv`, Node.js 22/npm, .NET 8 or 10, JDK 21 and Gradle 8.14.3. Maven and Java 17+ are used by Java examples/templates. The Java SDK build uses JDK 21 and emits Java 11 bytecode. The .NET SDK targets net6.0; templates target net8.0.

Building every SDK requires every listed language toolchain. Working on one language only needs its toolchain plus Go, Pulumi and the Python stamping scripts. See the [Makefile](Makefile), [component Makefile](components/Makefile) and [CI workflow](.github/workflows/ci.yml) for exact commands.

## Build and unit tests

From the repository root:

```sh
make provider components
make test
make lint
make check_schema check_components_schema
```

`make provider` regenerates the primitive schema and builds the provider plugin. `make components` builds the component plugin. `make test` runs provider and component Go tests; `make lint` runs provider `go vet`. Schema checks compare generated schemas to their committed snapshots.

Install local plugins when testing local programs:

```sh
make install_plugins
```

This installs both plugins from `bin/` into the current Pulumi plugin cache.

## Regenerate and build SDKs

Generated SDK sources are committed. Change the source mapping/component, regenerate the schema and affected SDKs, and include the generated changes in your pull request. Do not hand-edit generated resource classes.

```sh
make tfgen components_schema
make build_sdks build_app_sdks
```

For one language (`python`, `go`, `dotnet`, `java` or `nodejs`):

```sh
make build_python build_app_python
# Generation without packaging:
make generate_python
make -C components generate_python
```

The root build targets regenerate SDKs before compiling/packaging them. `components/Makefile` also exposes `gen_sdks` and `build_sdks`. Keep dependencies and all package identities consistent at the selected `VERSION`:

Component SDK generation first builds and installs the matching local `zerops` plugin into the current `PULUMI_HOME`; no pre-populated plugin cache is required. The component schema also records the primitive package version and GitHub plugin download URL, so external resource references retain their plugin source when resolved outside the monorepo. YAML examples and starters pin both the plugin version and download URL explicitly.

```sh
make VERSION=0.1.0 provider components build_sdks build_app_sdks
```

The Go `replace` directives in `components/go.mod` and `sdk-app/go.mod` point to `../sdk` for monorepo development; dependency consumers do not inherit them. Templates must reference published module/package identities, with no local replacements. Go releases require both `sdk/vX.Y.Z` and `sdk-app/vX.Y.Z` tags.

After changes, run the relevant CI equivalents:

```sh
python scripts/validate-templates.py
bash scripts/compile-example.sh python
```

Replace `python` with `go`, `dotnet`, `java` or `nodejs` as appropriate; build that SDK first. The template validator requires PyYAML. CI compiles each language and checks that regeneration leaves committed sources unchanged.

## Live acceptance tests

Live tests create billable Zerops resources and are **maintainer-run**. They require a dedicated Zerops development/test organization, a restricted token and the appropriate endpoint/organization configuration. Do not point them at a production organization or supply credentials to untrusted pull requests.

The [live workflow](.github/workflows/live.yml) runs only in the trusted repository/context, on its designated runner with environment-protected secrets. It builds local dependencies and invokes:

```sh
bash scripts/live-lifecycle.sh yaml-webapp
```

Maintainers configure the workflow environment and runner rather than copying credential values into programs. Credential-gated Go acceptance tests live in `examples/`; unit tests do not require an API token. A complete live check covers initial deployment and HTTP response, no-change preview, app update/redeployment, destroy and API-confirmed absence. Always clean up resources and state, including after failures; never publish raw logs containing credentials or IDs.

## Releasing

Only maintainers release. Prepare and review versioned schemas, SDKs, templates and [changelog](CHANGELOG.md), then create a `vX.Y.Z` tag on `main`. The [release workflow](.github/workflows/release.yml) verifies ancestry and metadata, tests the provider/components, builds both plugins with GoReleaser, creates the GitHub release, and adds `sdk/vX.Y.Z` and `sdk-app/vX.Y.Z` Go module tags.

Language registry jobs are opt-in through repository variables set to the string `true`:

| Variable | Protected environment | Authentication |
|---|---|---|
| `PUBLISH_PYPI` | `pypi` | PyPI trusted publishing (OIDC) |
| `PUBLISH_NPM` | `npm` | npm trusted publishing (OIDC), with provenance |
| `PUBLISH_NUGET` | `nuget` | NuGet trusted publishing/login (OIDC), configured `NUGET_USER` |
| `PUBLISH_MAVEN` | `maven-central` | Central Portal token and GPG signing secrets |

Each component registry job follows the primitive package job. Maven uploads signed artifacts through the Central staging endpoint; a maintainer must publish each deployment in the Central Portal. Maven is not an OIDC trusted-publishing job. Leave variables disabled until registry ownership, environments and credentials are configured. No registry credentials belong in the repository.

To publish packages after an existing GitHub release (for example, after configuring registry ownership and trusted publishers), enable the relevant `PUBLISH_*` variables and manually run **Release** from the Actions tab on `main`, supplying the existing `tag` such as `v0.1.0`. The workflow validates the `vX.Y.Z` input, checks out that tag, and repeats ancestry, version and unit-test verification. Registry jobs publish that same tag, retaining their environment gates and primitive-before-component ordering. A manual run does **not** recreate the GitHub release or Go module tags. Only enable registries whose versions have not already been published; published versions generally cannot be overwritten.
