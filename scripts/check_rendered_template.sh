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
    if (cd "$repo_dir" && rg -n --hidden --glob '!.git/**' --glob '!scripts/check_rendered_template.sh' "$pattern" .); then
      echo "ERROR: found stale $label" >&2
      exit 1
    fi
  else
    if (cd "$repo_dir" && grep -RInE --exclude-dir=.git --exclude=check_rendered_template.sh "$pattern" .); then
      echo "ERROR: found stale $label" >&2
      exit 1
    fi
  fi
}

scan_fixed() {
  local pattern="$1"
  local label="$2"

  if command -v rg >/dev/null 2>&1; then
    if (cd "$repo_dir" && rg -n --hidden --glob '!.git/**' --glob '!scripts/check_rendered_template.sh' --fixed-strings "$pattern" .); then
      echo "ERROR: found stale $label" >&2
      exit 1
    fi
  else
    if (cd "$repo_dir" && grep -RInF --exclude-dir=.git --exclude=check_rendered_template.sh "$pattern" .); then
      echo "ERROR: found stale $label" >&2
      exit 1
    fi
  fi
}

scan_regex '\{\{MODULE_NAME\}\}|\{\{MODULE_PATH\}\}|\{\{PACKAGE_NAME\}\}' "template placeholder"
if [[ "$module_path" != "github.com/ZoneCNH/configx" ]]; then
  scan_fixed "github.com/ZoneCNH/configx" "module path"
fi

if [[ "$module_name" != "baselibx" ]]; then
  scan_fixed "baselibx" "legacy smoke module name"
fi

if [[ "$module_name" != "corekit" ]]; then
  scan_fixed "corekit" "legacy smoke module name"
fi

scan_fixed "baselib-template" "legacy standard name"
scan_fixed "templatex" "legacy template module name"

if [[ "$module_name" != "configx" ]]; then
  scan_fixed "configx" "module name"
fi

if [[ "$package_name" != "configx" ]]; then
  scan_regex '\bconfigx\b' "package name"
fi

if [[ -f "$repo_dir/xlib-standard.lock" ]]; then
  if ! grep -Fq "module_name: \"$module_name\"" "$repo_dir/xlib-standard.lock"; then
    echo "ERROR: xlib-standard.lock module_name mismatch" >&2
    exit 1
  fi
  if ! grep -Fq "module_path: \"$module_path\"" "$repo_dir/xlib-standard.lock"; then
    echo "ERROR: xlib-standard.lock module_path mismatch" >&2
    exit 1
  fi
  if ! grep -Fq "package_name: \"$package_name\"" "$repo_dir/xlib-standard.lock"; then
    echo "ERROR: xlib-standard.lock package_name mismatch" >&2
    exit 1
  fi
fi

echo "rendered template check passed: $module_name"
