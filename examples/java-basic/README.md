# Zerops Java example

Deploy a tiny Node.js HTTP application using Java infrastructure code. Requires
JDK 17+, Maven, Pulumi CLI, the Pulumi Java language plugin, and a Zerops account/token.
The program requires `appName` (the explicit Zerops project name); optional
`message` defaults to `Hello from Zerops` and becomes the service's `MESSAGE` variable.

```sh
mvn -q package
# Choose one backend:
pulumi login                         # Pulumi Cloud
# pulumi login file://$HOME/pulumi-state  # local file backend
pulumi stack init dev
pulumi config set --secret zerops:token YOUR_TOKEN
pulumi config set appName my-java-app
pulumi config set message 'Hello from Java' # optional
pulumi up
curl "$(pulumi stack output url)"
# Edit app/server.js, then pulumi up to deploy the new source.
pulumi destroy
pulumi stack rm dev
```

`app/zerops.yml` is read verbatim; deployment waits for the message environment
variable, and public subdomain access waits for deployment. Outputs are
`projectId`, `serviceId`, and `url`.

## Local SDK development

The Maven dependency uses the published identity `io.zerops:pulumi:0.1.1`.
From the repository root, with JDK 17+, Gradle, Maven, and Pulumi CLI:

```sh
make build_java
gradle -p sdk/java publishToMavenLocal -x javadoc -x javadocJar
cd examples/java-basic
mvn -q package
```

The credential-gated integration test reads `ZEROPS_TOKEN` and optional
`ZEROPS_ENDPOINT`/`ZEROPS_ORGANIZATION_ID` values from the environment. Run it from
`examples/` with `go test -run '^TestJavaBasic$' -timeout 60m`.
