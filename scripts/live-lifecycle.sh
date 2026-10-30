#!/usr/bin/env bash
set -euo pipefail
set +x
root=$(cd "$(dirname "$0")/.." && pwd)
source "$root/scripts/live-common.sh"
example=${1:?Usage: live-lifecycle.sh example-directory}
[[ -d $example ]] || example="$root/examples/$example"
[[ -f $example/Pulumi.yaml ]] || { echo 'Example manifest missing' >&2; exit 2; }
state_root=${RUNNER_TEMP:-${TMPDIR:-/tmp}}
[[ ! -d /dev/shm ]] || state_root=/dev/shm
base=${LIVE_STATE_DIR:-$(mktemp -d "$state_root/pulacc-live-XXXXXX")}
base=$(realpath -m "$base")
[[ $base == "$state_root"/pulacc-live-* ]] || { echo 'Refusing state outside isolated live storage' >&2; exit 1; }
mkdir -p "$base/program" "$base/backend"
chmod 700 "$base"
export PULUMI_BACKEND_URL="file://$base/backend"
export PULUMI_CONFIG_PASSPHRASE
PULUMI_CONFIG_PASSPHRASE=$(openssl rand -hex 32)
mask "$PULUMI_CONFIG_PASSPHRASE"
cp -a "$example/." "$base/program/"
cd "$base/program"
printf '%s' "$PULUMI_CONFIG_PASSPHRASE" > "$base/passphrase"
chmod 600 "$base/passphrase"
name="pulacc-ci-${GITHUB_RUN_ID:-$(date +%s)}-$(openssl rand -hex 3)"
stack=$name
printf '%s\n' "$name" > "$base/project-name"
printf '%s\n' "$stack" > "$base/stack-name"
cleanup() {
  local result=$?
  trap - EXIT
  if [[ -f $base/initialized ]]; then
    if pulumi destroy --yes --non-interactive --stack "$stack"; then
      confirm_absent "$name" || result=1
      pulumi stack rm --yes --non-interactive "$stack" || result=1
    else
      result=1
    fi
  fi
  # Never retain local backend state or its encryption key, including on failure.
  rm -rf "$base"
  exit "$result"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
case "$(basename "$example")" in
  python-basic)
    python -m venv venv
    venv/bin/pip install "$root"/sdk/python/dist/*.whl
    pulumi_runtime_python=venv/bin/python
    export PULUMI_PYTHON_CMD="$pulumi_runtime_python"
    ;;
  go-basic)
    go mod edit -replace "github.com/zeropsio/pulumi-zerops/sdk=$root/sdk"
    go mod tidy
    ;;
  csharp-basic)
    dotnet restore --source "$root/sdk/dotnet/bin/Release" --source https://api.nuget.org/v3/index.json
    ;;
  java-basic)
    export MAVEN_ARGS="-Dmaven.repo.local=$base/m2"
    gradle -p "$root/sdk/java" --no-daemon "-Dmaven.repo.local=$base/m2" publishToMavenLocal
    mvn -B compile
    ;;
  ts-basic)
    tgz=$(cd "$root/sdk/nodejs/bin" && npm pack --silent --pack-destination "$base")
    npm install "$base/$tgz"
    ;;
esac
pulumi stack init "$stack" --non-interactive
touch "$base/initialized"
pulumi config set appName "$name"
pulumi config set projectName "$name"
pulumi config set message 'Hello from Zerops'
http_check() {
  local url deadline code
  url=$(pulumi stack output --json | jq -r '.url // empty')
  [[ -n $url ]] || return 0
  mask "$url"
  deadline=$((SECONDS + 600))
  while (( SECONDS < deadline )); do
    code=$(curl --silent --output /dev/null --write-out '%{http_code}' --max-time 20 "$url") || code=000
    if [[ $code == 200 ]]; then echo 'HTTP 200 confirmed'; return 0; fi
    sleep 10
  done
  echo 'HTTP readiness timed out' >&2
  return 1
}
pulumi up --yes --non-interactive
http_check
pulumi preview --expect-no-changes --non-interactive
if [[ -d app ]]; then
  python - <<'PY'
from pathlib import Path
source = Path('app/server.js')
if source.exists():
    text = source.read_text()
    if 'v1' not in text:
        raise SystemExit('Expected v1 application marker missing')
    source.write_text(text.replace('v1', 'v2'))
else:
    candidates = sorted(p for p in Path('app').rglob('*') if p.suffix == '.py')
    if not candidates:
        raise SystemExit('No supported application source to update')
    with candidates[0].open('a') as stream:
        stream.write('\n# pulacc lifecycle update\n')
PY
else
  pulumi config set serviceHostname appupdated
fi
pulumi up --yes --non-interactive
http_check
pulumi destroy --yes --non-interactive
confirm_absent "$name"
pulumi stack rm --yes --non-interactive "$stack"
rm "$base/initialized"
echo 'Lifecycle complete: create, stable preview, update, destroy, API absence'
