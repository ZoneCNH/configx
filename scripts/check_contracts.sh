#!/usr/bin/env bash
set -euo pipefail

required=(
  README.md
  docs/goal.md
  docs/contracts.md
  docs/current-state.md
  docs/examples.md
  docs/redaction.md
  docs/security.md
  docs/sources.md
  docs/envfile.md
  docs/merge.md
  docs/decode.md
  docs/validation.md
  docs/secrets.md
  docs/sanitize.md
  docs/xgo-integration.md
  docs/adr/ADR-20260601-001-explicit-source-loading.md
  docs/adr/ADR-20260601-002-no-global-config.md
  docs/adr/ADR-20260601-003-secret-handling.md
  docs/adr/ADR-20260601-004-yaml-toml-scope.md
  docs/release.md
  docs/review.md
  contracts/config.schema.json
  contracts/error.schema.json
  contracts/health.schema.json
  contracts/version.schema.json
  contracts/metrics.md
  contracts/manifest.schema.json
  release/manifest/template.json
)

for file in "${required[@]}"; do
  if [[ ! -s "$file" ]]; then
    echo "ERROR: missing or empty required contract artifact: $file" >&2
    exit 1
  fi
done

python3 -m json.tool contracts/config.schema.json >/dev/null
python3 -m json.tool contracts/error.schema.json >/dev/null
python3 -m json.tool contracts/health.schema.json >/dev/null
python3 -m json.tool contracts/version.schema.json >/dev/null
python3 -m json.tool contracts/manifest.schema.json >/dev/null
python3 -m json.tool release/manifest/template.json >/dev/null

grep -q 'github.com/ZoneCNH/configx' README.md docs/contracts.md release/manifest/template.json
grep -q 'docs/goal.md' README.md

GOWORK="${GOWORK:-off}" go test ./contracts

echo "contract artifacts OK"
