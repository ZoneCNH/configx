#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  scripts/render_template.sh --module-name NAME --module-path PATH --package-name NAME --out DIR [--enable-governance --layer LAYER --standard-version VERSION --standard-commit COMMIT]

Renders configx into a concrete base library by copying the repository, moving
pkg/configx to pkg/<package>, and replacing downstream identifiers.

When --enable-governance is supplied, the render must include the complete
xlib-standard governance pack and writes xlib-standard.lock provenance.
USAGE
}

module_name=""
module_path=""
package_name=""
out_dir=""
enable_governance=0
layer=""
standard_version=""
standard_commit=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --module-name)
      module_name="${2:-}"
      shift 2
      ;;
    --module-path)
      module_path="${2:-}"
      shift 2
      ;;
    --package-name)
      package_name="${2:-}"
      shift 2
      ;;
    --out)
      out_dir="${2:-}"
      shift 2
      ;;
    --enable-governance)
      enable_governance=1
      shift
      ;;
    --layer)
      layer="${2:-}"
      shift 2
      ;;
    --standard-version)
      standard_version="${2:-}"
      shift 2
      ;;
    --standard-commit)
      standard_commit="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "ERROR: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ -z "$module_name" || -z "$module_path" || -z "$package_name" || -z "$out_dir" ]]; then
  echo "ERROR: --module-name, --module-path, --package-name and --out are required" >&2
  usage >&2
  exit 2
fi

if [[ "$package_name" =~ [^a-zA-Z0-9_] || "$package_name" =~ ^[0-9] ]]; then
  echo "ERROR: --package-name must be a valid Go package identifier" >&2
  exit 2
fi

if [[ "$enable_governance" -eq 1 && ( -z "$layer" || -z "$standard_version" || -z "$standard_commit" ) ]]; then
  echo "ERROR: --enable-governance requires --layer, --standard-version and --standard-commit" >&2
  usage >&2
  exit 2
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
repo_abs="$(realpath "$repo_root")"
out_abs="$(realpath -m "$out_dir")"

if [[ "$out_abs" == "$repo_abs" || "$out_abs" == "$repo_abs"/* ]]; then
  echo "ERROR: output directory must be outside the source repository: $out_dir" >&2
  exit 2
fi

mkdir -p "$out_abs"
if find "$out_abs" -mindepth 1 -maxdepth 1 | read -r _; then
  echo "ERROR: output directory must be empty: $out_abs" >&2
  exit 2
fi
out_dir="$out_abs"

(
  cd "$repo_root"
  git ls-files -z -- \
    . \
    ':!release/manifest/latest.json' \
    ':!release/manifest/latest.json.sha256' \
    ':!release/standard-impact/latest.md' \
    ':!release/downstream-sync/latest.md' \
    ':!xlib-standard.lock' |
    tar --null -T - -cf -
) | (
  cd "$out_dir"
  tar -xf -
)

if [[ "$package_name" != "configx" ]]; then
  mkdir -p "$out_dir/pkg"
  mv "$out_dir/pkg/configx" "$out_dir/pkg/$package_name"
fi

collect_text_files() {
  find "$out_dir" -type f \( \
    -name '*.go' -o \
    -name '*.md' -o \
    -name '*.json' -o \
    -name '*.sh' -o \
    -name '*.yml' -o \
    -name '*.yaml' -o \
    -name 'Makefile' -o \
    -name 'go.mod' -o \
    -name '*.lock' \
  \) -print0
}

mapfile -d '' render_text_files < <(collect_text_files)

replace_in_text_files() {
  local find_text="$1"
  local replace_text="$2"

  if [[ "${#render_text_files[@]}" -eq 0 ]]; then
    return 0
  fi

  FIND_TEXT="$find_text" REPLACE_TEXT="$replace_text" perl -0pi -e 's/\Q$ENV{FIND_TEXT}\E/$ENV{REPLACE_TEXT}/g' "${render_text_files[@]}"
}

replace_in_text_files 'github.com/ZoneCNH/configx' "$module_path"
replace_in_text_files 'configx' "$module_name"
replace_in_text_files 'configx' "$package_name"

write_governance_lock() {
  cat > "$out_dir/xlib-standard.lock" <<EOF
schema_version: "1.0"
standard_name: "xlib-standard"
standard_repo: "https://github.com/ZoneCNH/xlib-standard"
standard_version: "$standard_version"
standard_commit: "$standard_commit"
module_name: "$module_name"
module_path: "$module_path"
package_name: "$package_name"
layer: "$layer"
adoption_status: "rendered"
adoption_check: "GOWORK=off make integration && GOWORK=off make contracts && GOWORK=off make boundary"
EOF
}

verify_governance_pack() {
  local required_paths=(
    ".githooks/pre-commit"
    ".githooks/pre-push"
    ".github/workflows/adoption-check.yml"
    ".github/rulesets/protect-main.json"
    ".agent/harness/harness.yaml"
    "mk/governance.mk"
  )

  local missing=0
  for path in "${required_paths[@]}"; do
    if [[ ! -e "$out_dir/$path" ]]; then
      echo "ERROR: --enable-governance requested but governance path is missing: $path" >&2
      missing=1
    fi
  done

  if [[ "$missing" -ne 0 ]]; then
    exit 1
  fi
}

if [[ "$enable_governance" -eq 1 ]]; then
  verify_governance_pack
  write_governance_lock
fi

(
  cd "$out_dir"
  gofmt -w ./pkg ./internal ./contracts ./examples ./testkit
)

echo "rendered $module_name at $out_dir"
