# Go: deploy an application to Zerops

Requires Go 1.24+, Pulumi CLI, a Zerops account/token, and the Zerops provider plugin 0.1.0. The bundled app is a dependency-free Node.js 22 HTTP server; Node.js runs on Zerops, not on your machine.

This example requires the published module `github.com/zeropsio/pulumi-zerops/sdk v0.1.0`. Its `go.mod` currently includes a local-development override:

```go
replace github.com/zeropsio/pulumi-zerops/sdk => ../../sdk
```

Contributors can run `make build_go` from the repository root to build the local SDK. The `replace` directive above makes this example use `../../sdk`; run `go build .` from this directory.

Choose Pulumi Cloud with `pulumi login`, or a local backend with `pulumi login file://$PWD/.state` (set `PULUMI_CONFIG_PASSPHRASE` for local state encryption). Then run:

```sh
pulumi stack init dev
pulumi config set --secret zerops:token "$ZEROPS_TOKEN"
pulumi config set appName my-app
# Optional; defaults to Hello from Zerops:
pulumi config set message 'Hello from my Go application'
pulumi up
curl "$(pulumi stack output url)"
```

`appName` is required and is the explicit Zerops project name. The program creates a LIGHT project, a Node.js service, its `MESSAGE` environment variable, a deployment from `./app`, and public subdomain access. It reads `app/zerops.yml` verbatim. Outputs are `projectId`, `serviceId`, and `url`.

Change `app/server.js` (for example, `v1` to `v2`) and run `pulumi up` to replace the deployment with the updated source. An unchanged `pulumi preview --expect-no-changes` should report no changes.

Clean up all resources and local stack state:

```sh
pulumi destroy
pulumi stack rm dev
```

The credential-gated integration test reads `ZEROPS_TOKEN` and optional `ZEROPS_ENDPOINT`/`ZEROPS_ORGANIZATION_ID` values from the environment. Run from `examples/` with `go test -run '^TestGoBasic$' -timeout 60m`.
