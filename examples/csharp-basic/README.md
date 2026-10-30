# Zerops C# example

Prerequisites: .NET 8 SDK, Pulumi CLI, and a Zerops integration token with project permissions. The program uses the published `Zerops.Pulumi` NuGet package at version `0.1.1`.

```sh
# Choose Pulumi Cloud, or a local backend instead:
pulumi login
# pulumi login file://$(pwd)/../../.pulumi-state
pulumi stack init dev
pulumi config set --secret zerops:token YOUR_TOKEN
pulumi config set appName my-csharp-app
# Optional; defaults to "Hello from Zerops":
pulumi config set message "Hello from C#"
pulumi up
curl "$(pulumi stack output url)"
```

`appName` explicitly names the LIGHT Zerops project. The program provisions a Node.js 22 service, sets its `MESSAGE` environment variable, deploys `app/`, and enables subdomain access. It exports `projectId`, `serviceId`, and `url`. Edit `app/server.js` and run `pulumi up` to deploy a new version.

Clean up:

```sh
pulumi destroy
pulumi stack rm dev
```

## Local SDK development

From the repository root, pack and restore the local SDK without changing the published package reference or adding machine-specific NuGet configuration:

```sh
dotnet pack sdk/dotnet -c Release -o .cache/nuget-local -p:PackageVersion=0.1.1
dotnet restore examples/csharp-basic \
  --source "$PWD/.cache/nuget-local" --source https://api.nuget.org/v3/index.json
dotnet build examples/csharp-basic --no-restore
```

The integration test packs the SDK and supplies a local NuGet source only in its temporary project copy. With development-environment credentials available, run `go test -run '^TestCSharpBasic$' -timeout 60m ./...` from `examples/`. No local absolute paths or token files belong in this example.
