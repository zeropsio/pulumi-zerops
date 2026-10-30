#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/live-common.sh"
listing=$(projects)
while IFS= read -r candidate; do
  [[ -n $candidate ]] || continue
  id=$(jq -r '.id' <<< "$candidate")
  [[ $id =~ ^[A-Za-z0-9_-]+$ ]] || exit 1
  # Re-read each project before deletion, refusing foreign ownership or fresh resources.
  project=$(api "${ZEROPS_ENDPOINT%/}/api/rest/public/project/$id")
  if jq -e --arg org "$ZEROPS_ORGANIZATION_ID" --argjson cutoff "$(($(date +%s) - 7200))" \
    '.clientId == $org and (.name|startswith("pulacc-ci-")) and ((.created|sub("\\.[0-9]+Z$";"Z")|fromdateiso8601) < $cutoff)' <<< "$project" >/dev/null; then
    name=$(jq -r '.name' <<< "$project")
    api -X DELETE "${ZEROPS_ENDPOINT%/}/api/rest/public/project/$id" >/dev/null
    confirm_absent "$name"
  fi
done < <(jq -c --arg org "$ZEROPS_ORGANIZATION_ID" '. | select(.clientId == $org and (.name|startswith("pulacc-ci-")))' <<< "$listing")
