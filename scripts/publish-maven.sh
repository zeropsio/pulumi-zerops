#!/usr/bin/env bash
# Compatibility endpoint docs: https://central.sonatype.org/publish/publish-portal-ossrh-staging-api/
set -euo pipefail
set +x
package=${1:?SDK directory required}
if [[ $package == sdk-app ]]; then
  gradle -p sdk/java --no-daemon publishToMavenLocal
fi
gradle -p "$package/java" --no-daemon publish
# Maven-like uploads must be transferred from the same runner/IP to the Portal.
auth=$(printf '%s:%s' "$PUBLISH_REPO_USERNAME" "$PUBLISH_REPO_PASSWORD" | base64 -w0)
printf '::add-mask::%s\n' "$auth"
curl --silent --show-error --fail-with-body -X POST \
  -H "Authorization: Bearer $auth" \
  'https://ossrh-staging-api.central.sonatype.com/manual/upload/defaultRepository/io.zerops?publishing_type=user_managed'
# A maintainer now reviews and clicks Publish in https://central.sonatype.com/publishing.
