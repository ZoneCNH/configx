#!/usr/bin/env bash
set -euo pipefail

args=(--verify release/manifest/latest.json)

if [[ "${RELEASE_EVIDENCE_REQUIRE_PASSED:-0}" == "1" ]]; then
  args+=(--require-passed)
fi

if [[ "${RELEASE_EVIDENCE_REQUIRE_CLEAN:-0}" == "1" ]]; then
  args+=(--require-clean)
fi

if [[ -n "${VERSION:-}" ]]; then
  args+=(--expect-version "$VERSION")
fi

go run ./internal/tools/releasemanifest "${args[@]}"

artifact=release/manifest/latest.json
for forbidden in \
  '/home/k8s/secrets/env' \
  '.env' \
  'production.yaml' \
  'production.yml' \
  'config.local.yaml' \
  'config.local.yml'; do
  if grep -Fq "$forbidden" "$artifact"; then
    echo "ERROR: release evidence contains forbidden config discovery literal: $forbidden" >&2
    exit 1
  fi
done

if grep -Eiq '(password|passwd|token|access_key|secret_key)[[:space:]]*[:=][[:space:]]*["'"'']?[^"'"'',}[:space:]]{8,}' "$artifact"; then
  echo "ERROR: release evidence contains possible raw secret material" >&2
  exit 1
fi
