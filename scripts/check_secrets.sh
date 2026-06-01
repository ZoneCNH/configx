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

for pattern in "${patterns[@]}"; do
  if grep -R -n -E "$pattern" . \
    --exclude-dir=.git \
    --exclude-dir=.omx \
    --exclude-dir=vendor \
    --exclude='*.sum' \
    --exclude='check_secrets.sh' \
    --exclude='goal.md'; then
    echo "ERROR: possible secret matched pattern: $pattern" >&2
    exit 1
  fi
done

echo "secret scan OK"
