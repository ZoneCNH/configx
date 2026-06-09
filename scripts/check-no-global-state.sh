#!/usr/bin/env bash
set -euo pipefail

# check-no-global-state.sh
#
# CI script: checks that production code does not contain process-level
# configuration singletons or global config state initialization.
#
# Patterns detected:
#   1. var.*=.*New(           — package-level variable initialized by constructor
#   2. sync.Once.*Config      — sync.Once used to initialize config
#   3. init\(\).*Config       — init() function that touches config
#
# Exit 0 if clean, exit 1 with violation locations if any pattern is found.

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

SEARCH_DIRS=(pkg internal)

# Collect all Go source files excluding test files and vendor.
mapfile -t go_files < <(
  find "${SEARCH_DIRS[@]}" -name '*.go' \
    ! -name '*_test.go' \
    ! -path '*/vendor/*' \
    ! -path '*/.git/*' 2>/dev/null
)

if [[ ${#go_files[@]} -eq 0 ]]; then
  echo "no Go source files found in ${SEARCH_DIRS[*]}"
  exit 0
fi

violations=0

echo "checking for global config state patterns..."

# Pattern 1: Package-level var initialized by constructor
# Matches: var X = New... or var X = configx.New...
while IFS= read -r match; do
  echo "VIOLATION: package-level config singleton: $match"
  ((violations++))
done < <(
  grep -n -E '^\s*var\s+[[:alnum:]_]+[[:space:]]*=?[[:space:]]*[[:alnum:]_.]*New[[:alnum:]_]*\(' "${go_files[@]}" 2>/dev/null || true
)

# Pattern 2: sync.Once used with Config
while IFS= read -r match; do
  echo "VIOLATION: sync.Once config singleton: $match"
  ((violations++))
done < <(
  grep -n -E 'sync\.Once.*[Cc]onfig|[Cc]onfig.*sync\.Once' "${go_files[@]}" 2>/dev/null || true
)

# Pattern 3: init() function that references config
while IFS= read -r match; do
  # We check that init() itself or the block contains config references.
  echo "VIOLATION: init() config initialization: $match"
  ((violations++))
done < <(
  grep -n -E '^func init\(\)' "${go_files[@]}" 2>/dev/null \
    | grep -i -E 'config' || true
)

# Also check for the broader pattern: init function bodies that set package-level config.
# This is a heuristic: find init() declarations, then check the next 20 lines for config assignments.
while IFS= read -r file_line; do
  file="${file_line%%:*}"
  line="${file_line##*:}"
  # Extract the line number
  lineno=$(echo "$line" | grep -oE '^[0-9]+')
  if [[ -z "$lineno" ]]; then
    continue
  fi
  # Check next 20 lines for config-related assignments.
  end=$((lineno + 20))
  block=$(sed -n "${lineno},${end}p" "$file" 2>/dev/null)
  if echo "$block" | grep -q -i -E '\bconfig\b.*=|=\s*.*\bconfig\b'; then
    echo "VIOLATION: init() function assigns config variable: ${file}:${lineno}"
    ((violations++))
  fi
done < <(
  grep -n -H '^func init()' "${go_files[@]}" 2>/dev/null || true
)

# Pattern 4: Package-level singletons (var X *ConfigType)
while IFS= read -r match; do
  echo "VIOLATION: package-level config pointer: $match"
  ((violations++))
done < <(
  grep -n -E '^\s*var\s+[[:alnum:]_]+[[:space:]]+\*[[:alnum:]_]*[Cc]onfig[[:alnum:]_]*\b' "${go_files[@]}" 2>/dev/null || true
)

if [[ $violations -gt 0 ]]; then
  echo ""
  echo "FAIL: found $violations global state violation(s)"
  echo "configx must not create process-level config singletons."
  echo "All configuration loading must be explicit and caller-driven."
  exit 1
fi

echo "no global config state patterns found — OK"
