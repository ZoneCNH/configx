#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  scripts/check_rendered_template.sh DIR MODULE_NAME MODULE_PATH PACKAGE_NAME

Checks that a rendered template has no stale template identifiers and exposes
the expected Go module and package directory.
USAGE
}

if [[ $# -ne 4 ]]; then
  usage >&2
  exit 2
fi

repo_dir="$1"
module_name="$2"
module_path="$3"
package_name="$4"

if [[ ! -d "$repo_dir" ]]; then
  echo "ERROR: rendered directory does not exist: $repo_dir" >&2
  exit 2
fi

actual_module="$(cd "$repo_dir" && GOWORK=off go list -m)"
if [[ "$actual_module" != "$module_path" ]]; then
  echo "ERROR: module path mismatch: got $actual_module, want $module_path" >&2
  exit 1
fi

if [[ ! -d "$repo_dir/pkg/$package_name" ]]; then
  echo "ERROR: rendered package directory missing: pkg/$package_name" >&2
  exit 1
fi

if [[ "$package_name" != "configx" && -e "$repo_dir/pkg/configx" ]]; then
  echo "ERROR: stale pkg/configx directory still exists" >&2
  exit 1
fi

scan_regex() {
  local pattern="$1"
  local label="$2"

  if command -v rg >/dev/null 2>&1; then
    if rg -n --hidden --glob '!.git/**' "$pattern" "$repo_dir"; then
      echo "ERROR: found stale $label" >&2
      exit 1
    fi
  else
    if grep -RInE --exclude-dir=.git "$pattern" "$repo_dir"; then
      echo "ERROR: found stale $label" >&2
      exit 1
    fi
  fi
}

scan_fixed() {
  local pattern="$1"
  local label="$2"

  if command -v rg >/dev/null 2>&1; then
    if rg -n --hidden --glob '!.git/**' --fixed-strings "$pattern" "$repo_dir"; then
      echo "ERROR: found stale $label" >&2
      exit 1
    fi
  else
    if grep -RInF --exclude-dir=.git "$pattern" "$repo_dir"; then
      echo "ERROR: found stale $label" >&2
      exit 1
    fi
  fi
}

scan_regex '\{\{MODULE_NAME\}\}|\{\{MODULE_PATH\}\}|\{\{PACKAGE_NAME\}\}' "template placeholder"
scan_fixed "github.com/ZoneCNH/configx" "module path"

if [[ "$module_name" != "configx" ]]; then
  scan_fixed "configx" "module name"
fi

if [[ "$package_name" != "configx" ]]; then
  scan_regex '\bconfigx\b' "package name"
fi

echo "rendered template check passed: $module_name"
