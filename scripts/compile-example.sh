#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "$0")/.." && pwd)
cd "$root"
case "$1" in
  go) (cd examples/go-basic && go mod edit -replace github.com/zeropsio/pulumi-zerops/sdk="$root/sdk" && go mod tidy && go vet ./...) ;;
  python)
    python -m venv .cache/example-python
    .cache/example-python/bin/pip install sdk/python/dist/*.whl
    .cache/example-python/bin/python -m compileall -q examples/python-basic
    ;;
  dotnet)
    dotnet restore examples/csharp-basic --source "$root/sdk/dotnet/bin/Release" --source https://api.nuget.org/v3/index.json
    dotnet build examples/csharp-basic --no-restore
    ;;
  java)
    gradle -p sdk/java --no-daemon publishToMavenLocal
    mvn -B -f examples/java-basic/pom.xml compile
    ;;
  nodejs)
    (cd examples/ts-basic && npm install --no-save "$root/sdk/nodejs/bin" && npm run build)
    ;;
  *) echo 'Unknown SDK language' >&2; exit 2 ;;
esac
