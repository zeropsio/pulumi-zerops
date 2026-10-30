#!/usr/bin/env bash
# Sourced by live helpers. Never enable shell tracing: inputs include credentials.
set -euo pipefail
set +x
: "${ZEROPS_ENDPOINT:?}" "${ZEROPS_TOKEN:?}" "${ZEROPS_ORGANIZATION_ID:?}" "${ZEROPS_DEV_DOMAIN:?}"
export ZEROPS_SANDBOX=dev
python - <<'PY'
import os
from urllib.parse import urlparse
host = urlparse(os.environ['ZEROPS_ENDPOINT']).hostname or ''
domain = os.environ['ZEROPS_DEV_DOMAIN'].strip('.').lower()
parent = domain.partition('.')[2]
if not parent or not (host.lower().endswith('.' + parent) or host.lower() == parent):
    raise SystemExit('Refusing endpoint outside the configured development domain parent')
PY
mask() { if [[ ${GITHUB_ACTIONS:-} == true ]]; then printf '::add-mask::%s\n' "$1"; fi; }
mask "$ZEROPS_ENDPOINT"; mask "$ZEROPS_TOKEN"; mask "$ZEROPS_ORGANIZATION_ID"; mask "$ZEROPS_DEV_DOMAIN"
api() {
  curl --silent --show-error --fail-with-body --connect-timeout 20 --max-time 120 \
    -H "Authorization: Bearer $ZEROPS_TOKEN" -H 'Content-Type: application/json' "$@"
}
projects() {
  local offset=0 response count total
  while :; do
    response=$(api -X POST "${ZEROPS_ENDPOINT%/}/api/rest/public/project/search" \
      -d "$(jq -nc --arg org "$ZEROPS_ORGANIZATION_ID" --argjson offset "$offset" '{search:[{name:"clientId",operator:"eq",value:$org}],limit:100,offset:$offset}')")
    jq -e '.items | type == "array"' <<< "$response" >/dev/null
    jq -c '.items[]' <<< "$response"
    count=$(jq '.items|length' <<< "$response")
    total=$(jq '.total // .count // 0' <<< "$response")
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
