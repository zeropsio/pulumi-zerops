# TypeScript: deploy a Node.js app

Requires Node.js 22+, npm, Pulumi CLI, and a Zerops account/token. The package
manifest uses the published `@zerops/pulumi` SDK at version `0.1.1`.

```sh
npm install
npm run build
pulumi login # Pulumi Cloud; alternatively: pulumi login file://$(pwd)/.pulumi-state
pulumi stack init dev
pulumi config set --secret zerops:token YOUR_TOKEN
pulumi config set appName my-app
# Optional; defaults to Hello from Zerops:
pulumi config set message 'Hello from TypeScript'
pulumi up
curl "$(pulumi stack output url)"
pulumi destroy
pulumi stack rm dev
```

`appName` explicitly names the Zerops project. The program creates a LIGHT
project, a Node.js 22 runtime, a MESSAGE environment variable, a deployment from
`app/`, and public subdomain access. Deployment waits for MESSAGE; public access
waits for deployment. Edit the version marker in `app/server.js` and run
`pulumi up` again to redeploy the application.

## Local SDK development

From the repository root, build and pack the local SDK, then install it in this
example:

```sh
make build_nodejs
npm pack ./sdk/nodejs/bin --pack-destination /tmp
cd examples/ts-basic
npm install --no-package-lock /tmp/zerops-pulumi-0.1.1.tgz
npm run build
```

The credential-gated integration test reads `ZEROPS_TOKEN` and optional
`ZEROPS_ENDPOINT`/`ZEROPS_ORGANIZATION_ID` values from the environment. Run from
`examples/` with `go test -run '^TestNodeJSBasic$' -timeout 60m -v`. It checks HTTP
responses before and after an application-source update, performs the no-change
lifecycle check, destroys its resources, and checks project absence.
