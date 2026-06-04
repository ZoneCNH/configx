#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

patterns=(
  'password[[:space:]]*='
  'passwd[[:space:]]*='
  'token[[:space:]]*='
  'access_key[[:space:]]*='
  'secret_key[[:space:]]*='
  "secret[[:space:]]*=[[:space:]]*[\"'][^\"']{8,}[\"']"
  'AKIA[0-9A-Z]{16}'
  'BEGIN RSA PRIVATE KEY'
  'BEGIN OPENSSH PRIVATE KEY'
  'BEGIN PRIVATE KEY'
)

files=()
while IFS= read -r -d '' file; do
  case "$file" in
    vendor/*|*.sum|scripts/check_secrets.sh|docs/goal.md)
      continue
      ;;
  esac

  if [[ -f "$file" ]]; then
    files+=("$file")
  fi
done < <(git ls-files -co --exclude-standard -z)

if ((${#files[@]} == 0)); then
  echo "secret scan OK"
  exit 0
fi

for pattern in "${patterns[@]}"; do
  if grep -n -I -E "$pattern" "${files[@]}"; then
    echo "ERROR: possible secret matched pattern: $pattern" >&2
    exit 1
  fi
done

echo "secret scan OK"
