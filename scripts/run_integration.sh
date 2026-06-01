#!/usr/bin/env bash
set -euo pipefail

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

cases=(
  "foundationx|github.com/ZoneCNH/foundationx|foundationx"
  "corekit|example.com/acme/corekit|corekit"
)

for spec in "${cases[@]}"; do
  IFS='|' read -r module_name module_path package_name <<< "$spec"
  out_dir="$tmpdir/$module_name"

  ./scripts/render_template.sh \
    --module-name "$module_name" \
    --module-path "$module_path" \
    --package-name "$package_name" \
    --out "$out_dir"

  ./scripts/check_rendered_template.sh "$out_dir" "$module_name" "$module_path" "$package_name"

  (
    cd "$out_dir"
    git init -q
    git config user.email "ci@example.invalid"
    git config user.name "Template Integration"
    git add .
    git commit -qm "Initial rendered template"

    GOWORK=off go test ./...
    GOWORK=off make contracts
    GOWORK=off make boundary
    CHECK_STATUS=passed GOWORK=off make evidence
    RELEASE_EVIDENCE_REQUIRE_PASSED=1 GOWORK=off make release-evidence-check
  )
done

echo "integration check passed"
