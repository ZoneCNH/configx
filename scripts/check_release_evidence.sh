#!/usr/bin/env bash
set -euo pipefail

args=(--verify release/manifest/latest.json)

if [[ "${RELEASE_EVIDENCE_REQUIRE_PASSED:-0}" == "1" ]]; then
  args+=(--require-passed)
fi

if [[ "${RELEASE_EVIDENCE_REQUIRE_CLEAN:-0}" == "1" ]]; then
  args+=(--require-clean)
fi

go run ./internal/tools/releasemanifest "${args[@]}"
