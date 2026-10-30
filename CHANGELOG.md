# Changelog

## 0.1.0 — 2026-10-29

First release.

- Bridged `zerops` provider 0.1.0, built from `terraform-provider-zerops` v0.1.1 with `pulumi-terraform-bridge` v3.142.0. Tested with Pulumi CLI 3.273.0.
- All 19 primitive resources and 9 data functions for Zerops projects, services, deployments, storage, routing and related configuration.
- Separate `zerops-app` 0.1.0 component package with `WebApp`: a LIGHT project, runtime service, directory deployment, public URL and optional single-node PostgreSQL.
- SDKs for Python, Go, .NET, Java and Node.js, plus schema-native YAML support for both packages.
- Versioned starter templates for Python, Go, C#, Java, TypeScript and YAML; complete examples and language quickstarts.
- Secret token configuration and server-side database connection references, without resolved database credential readback by WebApp.
