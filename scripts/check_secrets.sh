#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

high_confidence_patterns=(
  'AKIA[0-9A-Z]{16}'
  'BEGIN RSA PRIVATE KEY'
  'BEGIN OPENSSH PRIVATE KEY'
  'BEGIN PRIVATE KEY'
)

key_value_patterns=(
  '(^|[^[:alnum:]_])[[:alnum:]_]*(password|passwd|token|access_key|secret_key)[[:alnum:]_]*[[:space:]]*=[[:space:]]*["'\'']?[^"'\''[:space:]]{8,}'
  '(^|[^[:alnum:]_])[[:alnum:]_]*secret[[:alnum:]_]*[[:space:]]*=[[:space:]]*["'\''][^"'\''[:space:]]{8,}["'\'']'
)

files=()
key_value_files=()
while IFS= read -r -d '' file; do
  case "$file" in
    vendor/*|*.sum|scripts/check_secrets.sh|docs/goal.md)
      continue
      ;;
  esac

  if [[ -f "$file" ]]; then
    files+=("$file")
    case "$file" in
      docs/*|*_test.go|testdata/*|*/testdata/*)
        ;;
      *)
        key_value_files+=("$file")
        ;;
    esac
  fi
done < <(git ls-files -co --exclude-standard -z)

if ((${#files[@]} == 0)); then
  echo "secret scan OK"
  exit 0
fi

for pattern in "${high_confidence_patterns[@]}"; do
  if grep -n -I -E "$pattern" "${files[@]}"; then
    echo "ERROR: possible secret matched pattern: $pattern" >&2
    exit 1
  fi
done

if ((${#key_value_files[@]} > 0)); then
  for pattern in "${key_value_patterns[@]}"; do
    if grep -n -I -E "$pattern" "${key_value_files[@]}"; then
      echo "ERROR: possible secret matched pattern: $pattern" >&2
      exit 1
    fi
  done
fi

echo "secret scan OK"
