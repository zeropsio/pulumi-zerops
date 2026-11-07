# Changelog

## Unreleased

- Move live acceptance to the protected `prod-live` environment and `zerops-prod` runner, retaining secret-free ref/ancestry verification and SHA-pinned actions. Fail closed before tool installation on missing production configuration.
- Keep the released Terraform v0.1.2 production guard and its `tfacc-` ownership prefix/creation ledger. Use only the dedicated production CI organization, small LIGHT/single-node resources with no IPv4, and an always-run scoped sweep that verifies zero `tfacc-` leftovers.
- Route live sweeper deletions through the tagged provider's `zguard` ownership ledger and production transport, adopting only API-verified CI projects. Reserve a bounded cleanup window and leave fixed-topology PostgreSQL horizontal sizing unset while keeping its 2 GB disk minimum.

## 0.1.1 — 2026-10-30

- Pin `terraform-provider-zerops` v0.1.2 from the Go proxy, without a local replacement.
- Both `zerops` and `zerops-app`, their generated SDKs, component dependency and starter templates now use 0.1.1.
- **Breaking:** preview rejects unqualified dual-mode managed-service types in `Service.type` and known `ProjectImport` YAML: use explicit `:single@` or `:ha@` modes for postgresql, mariadb, valkey, clickhouse, elasticsearch, kafka, nats, qdrant, seaweedfs and typesense. WebApp defaults to `postgresql:single@16` and validates explicit PostgreSQL single/HA modes.
- **Upgrade:** for existing shorthand state, set `type` to the qualified value in its `base` output (for example, `postgresql:ha@16`). This updates state without replacement; selecting the other mode still replaces the service.
- Preview defaults and enforces ClickHouse RAM >= 0.375 GB (`ZEROPS_CLICKHOUSE_MIN_RAM_GB`) and Elasticsearch RAM >= 2 GB (`ZEROPS_ELASTICSEARCH_MIN_RAM_GB`). These are environment-only overrides; there are no new Pulumi provider config keys. PostgreSQL disk remains >= 2 GB by default.
- `verticalAutoscaling.minRamGb` and `maxRamGb` are now optional/computed, allowing provider-planned RAM defaults.

## 0.1.0 — 2026-10-29

First release.

- Bridged `zerops` provider 0.1.0, built from `terraform-provider-zerops` v0.1.1 with `pulumi-terraform-bridge` v3.142.0. Tested with Pulumi CLI 3.273.0.
- All 19 primitive resources and 9 data functions for Zerops projects, services, deployments, storage, routing and related configuration.
- Separate `zerops-app` 0.1.0 component package with `WebApp`: a LIGHT project, runtime service, directory deployment, public URL and optional single-node PostgreSQL.
- SDKs for Python, Go, .NET, Java and Node.js, plus schema-native YAML support for both packages.
- Versioned starter templates for Python, Go, C#, Java, TypeScript and YAML; complete examples and language quickstarts.
- Secret token configuration and server-side database connection references, without resolved database credential readback by WebApp.
