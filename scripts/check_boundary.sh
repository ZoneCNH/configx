#!/usr/bin/env bash
set -euo pipefail

echo "checking forbidden dependency on x.go..."

DEPS="$(GOWORK="${GOWORK:-off}" go list -deps ./...)"
FORBIDDEN_DEPS=(
  "github.com/bytechainx/x.go"
  "github.com/ZoneCNH/x.go"
)

for dep in "${FORBIDDEN_DEPS[@]}"; do
  if grep -Fq "$dep" <<<"$DEPS"; then
    echo "ERROR: base library template must not depend on x.go dependency: $dep"
    exit 1
  fi
done

echo "checking forbidden business terms..."

FORBIDDEN_TERMS=(
  "MacroRegime"
  "MarketRegime"
  "TradingSignal"
  "BTCUSDT"
  "ETHUSDT"
  "Kline"
  "OrderBook"
  "Position"
  "RiskGate"
)

for term in "${FORBIDDEN_TERMS[@]}"; do
  if grep -R --line-number --fixed-strings "$term" ./pkg ./internal --exclude-dir=.git; then
    echo "ERROR: forbidden business term found: $term"
    exit 1
  fi
done

echo "boundary check passed"
