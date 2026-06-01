#!/usr/bin/env bash
set -euo pipefail

required=(
  README.md
  docs/goal.md
  docs/contracts.md
  docs/examples.md
  docs/release.md
  docs/review.md
  contracts/config.schema.json
  contracts/error.schema.json
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
python3 -m json.tool contracts/manifest.schema.json >/dev/null
python3 -m json.tool release/manifest/template.json >/dev/null

grep -q 'github.com/bytechainx/configx' README.md docs/contracts.md release/manifest/template.json
grep -q 'docs/goal.md' README.md

echo "contract artifacts OK"
