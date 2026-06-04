#!/usr/bin/env bash
set -euo pipefail

echo "running fuzz smoke..."

found=0
fuzz_time="${FUZZ_SMOKE_TIME:-10s}"
packages=$(GOWORK="${GOWORK:-off}" go list ./...)

for pkg in $packages; do
	dir=$(GOWORK="${GOWORK:-off}" go list -f '{{.Dir}}' "$pkg")
	while IFS= read -r fuzz_name; do
		found=1
		GOWORK="${GOWORK:-off}" go test "$pkg" -run=^$ -fuzz="^${fuzz_name}$" -fuzztime="$fuzz_time"
	done < <(
		grep -R -h -E --include='*_test.go' '^func Fuzz[A-Za-z0-9_]*\(' "$dir" \
			| sed -E 's/^func (Fuzz[A-Za-z0-9_]*)\(.*$/\1/'
	)
done

if [[ "$found" -eq 0 ]]; then
	echo "no fuzz tests found"
fi

echo "fuzz smoke passed"
