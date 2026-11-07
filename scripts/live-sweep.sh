#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/live-common.sh"
[[ ${PULUMI_ZGUARD:-} == /* && -x ${PULUMI_ZGUARD:-} ]] || { echo 'An absolute executable PULUMI_ZGUARD path is required' >&2; exit 1; }
listing=$(projects)
while IFS= read -r candidate; do
  [[ -n $candidate ]] || continue
  id=$(jq -r '.id' <<< "$candidate")
  [[ $id =~ ^[A-Za-z0-9_-]+$ ]] || exit 1
  # Re-read immediately before deletion; never infer ownership from a hostname.
  project=$(api "$ZEROPS_ENDPOINT/api/rest/public/project/$id")
  jq -e --arg org "$ZEROPS_ORGANIZATION_ID" --arg id "$id" \
    '.id == $id and .clientId == $org and (.name|startswith("tfacc-"))' <<< "$project" >/dev/null || { echo 'Refusing cleanup after project ownership changed' >&2; exit 1; }
  name=$(jq -r '.name' <<< "$project")
  # Adopt only the exact API-verified CI project into the provider's ownership
  # ledger. The pinned guard rechecks API ownership through its production transport.
  "$PULUMI_ZGUARD" record-project "$id" "$name" probe
  response=$("$PULUMI_ZGUARD" request DELETE "/project/$id")
  jq -e '.status >= 200 and .status < 300' <<< "$response" >/dev/null || { echo 'Guarded project deletion failed' >&2; exit 1; }
  confirm_absent "$name"
done < <(jq -c --arg org "$ZEROPS_ORGANIZATION_ID" 'select(.clientId == $org and (.name|startswith("tfacc-")))' <<< "$listing")
# A successful sweep means zero CI projects, not merely zero aged candidates.
listing=$(projects)
if jq -s -e 'any(.[]; .name|startswith("tfacc-"))' <<< "$listing" >/dev/null; then
  echo 'Production cleanup failed: tfacc- projects remain in the CI organization' >&2
  exit 1
fi
echo 'Production sweep confirmed zero tfacc- projects in the CI organization'
