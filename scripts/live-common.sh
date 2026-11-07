#!/usr/bin/env bash
# Sourced by live helpers. Never enable tracing: inputs include credentials.
set -euo pipefail
set +x
for key in ZEROPS_ENDPOINT ZEROPS_TOKEN ZEROPS_ORGANIZATION_ID ZEROPS_CI_ORGANIZATION_ID; do
  [[ -n ${!key:-} && -n ${!key//[[:space:]]/} ]] || { echo "Missing required production secret: $key" >&2; exit 1; }
done
[[ $ZEROPS_ENDPOINT == https://api.app-prg1.zerops.io ]] || { echo 'Production endpoint must be https://api.app-prg1.zerops.io' >&2; exit 1; }
[[ ${ZEROPS_SANDBOX:-} == prod && ${ZEROPS_TARGET:-prod} == prod ]] || { echo 'Live acceptance requires ZEROPS_SANDBOX=prod' >&2; exit 1; }
[[ $ZEROPS_ORGANIZATION_ID =~ ^[A-Za-z0-9_-]+$ ]] || { echo 'Invalid CI organization ID' >&2; exit 1; }
[[ $ZEROPS_CI_ORGANIZATION_ID == "$ZEROPS_ORGANIZATION_ID" ]] || { echo 'Refusing organization outside the dedicated production CI organization' >&2; exit 1; }
export ZEROPS_TARGET=prod
export ZEROPS_PROD_LEDGER=${ZEROPS_PROD_LEDGER:-${RUNNER_TEMP:-${TMPDIR:-/tmp}}/pulacc-production-ledger.tsv}
[[ $ZEROPS_PROD_LEDGER == /* ]] || { echo 'Production ledger path must be absolute' >&2; exit 1; }
umask 077
mask() { if [[ ${GITHUB_ACTIONS:-} == true ]]; then printf '::add-mask::%s\n' "$1"; fi; }
mask "$ZEROPS_TOKEN"; mask "$ZEROPS_ORGANIZATION_ID"
api() {
  curl --silent --show-error --fail --connect-timeout 20 --max-time 120 \
    -H @- -H 'Content-Type: application/json' "$@" <<< "Authorization: Bearer $ZEROPS_TOKEN"
}
projects() {
  local offset=0 response count total
  while :; do
    response=$(api -X POST "$ZEROPS_ENDPOINT/api/rest/public/project/search" \
      -d "$(jq -nc --arg org "$ZEROPS_ORGANIZATION_ID" --argjson offset "$offset" '{search:[{name:"clientId",operator:"eq",value:$org}],limit:100,offset:$offset}')") || return 1
    jq -e --arg org "$ZEROPS_ORGANIZATION_ID" '.items | type == "array" and all(.[]; .clientId == $org and (.id | type == "string" and test("^[A-Za-z0-9_-]+$")) and (.name | type == "string"))' <<< "$response" >/dev/null || { echo 'Project search returned invalid or foreign organization data' >&2; return 1; }
    jq -c '.items[]' <<< "$response" || return 1
    count=$(jq '.items|length' <<< "$response") || return 1
    total=$(jq '.total // .count // 0' <<< "$response") || return 1
    offset=$((offset + count))
    if (( count == 0 || count < 100 && offset >= total )); then break; fi
  done
}
confirm_absent() {
  local name=$1 attempt found
  for attempt in {1..60}; do
    found=$(projects) || return 1
    if ! jq -s -e --arg name "$name" 'any(.[]; .name == $name)' <<< "$found" >/dev/null; then
      echo 'Project absence confirmed by API'
      return 0
    fi
    sleep 10
  done
  echo 'Project still present after destroy' >&2
  return 1
}
