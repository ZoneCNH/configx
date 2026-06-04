#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

echo "checking forbidden generated files..."

if found="$(find . \
  -path './.git' -prune -o \
  -path './.omx' -prune -o \
  -path './tmp' -prune -o \
  -name 'x.go' -print -quit)" && [[ -n "$found" ]]; then
  echo "ERROR: base library template must not contain generated x.go file: $found"
  exit 1
fi

echo "checking forbidden dependencies..."

DEPS="$(GOWORK="${GOWORK:-off}" go list -deps ./...)"
FORBIDDEN_DEPS=(
  "github.com/bytechainx/x.go"
  "github.com/ZoneCNH/x.go"
  "github.com/redis/go-redis"
  "github.com/IBM/sarama"
  "github.com/Shopify/sarama"
  "github.com/segmentio/kafka-go"
  "github.com/jackc/pgx"
  "github.com/lib/pq"
  "gorm.io/gorm"
  "github.com/taosdata/driver-go"
  "github.com/aws/aws-sdk-go"
  "github.com/aws/aws-sdk-go-v2"
  "github.com/aliyun/aliyun-oss-go-sdk"
)

for dep in "${FORBIDDEN_DEPS[@]}"; do
  if grep -Fq "$dep" <<<"$DEPS"; then
    echo "ERROR: base library template must not depend on forbidden infrastructure dependency: $dep"
    exit 1
  fi
done

SEARCH_DIRS=(pkg internal contracts examples)

echo "checking forbidden implicit config discovery..."

# Explicit config loading is allowed: callers may pass a path to LoadEnvFile,
# NewEnvFileSource, LoadYAMLFile, etc. The boundary this script enforces is
# implicit discovery/defaulting in production code, such as opening ".env" or
# production.yaml directly from library code.
FORBIDDEN_DISCOVERY_PATTERNS=(
  '(^|[^[:alnum:]_])((os\.)?(Open|ReadFile|Stat|Lstat)|filepath\.Abs|filepath\.Join)[[:space:]]*\([^)]*["'\'']\.env["'\'']'
  '(^|[^[:alnum:]_])((os\.)?(Open|ReadFile|Stat|Lstat)|filepath\.Abs|filepath\.Join)[[:space:]]*\([^)]*["'\'']production\.yaml["'\'']'
  '(^|[^[:alnum:]_])((os\.)?(Open|ReadFile|Stat|Lstat)|filepath\.Abs|filepath\.Join)[[:space:]]*\([^)]*["'\'']/home/k8s/secrets/env'
)

for pattern in "${FORBIDDEN_DISCOVERY_PATTERNS[@]}"; do
  if grep -R --line-number --extended-regexp "$pattern" "${SEARCH_DIRS[@]}" \
    --exclude='*_test.go' \
    --exclude-dir=.git; then
    echo "ERROR: forbidden implicit config discovery pattern found: $pattern"
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
  if [ "${#SEARCH_DIRS[@]}" -gt 0 ] && grep -R --line-number --fixed-strings "$term" "${SEARCH_DIRS[@]}" --exclude-dir=.git; then
    echo "ERROR: forbidden business term found: $term"
    exit 1
  fi
done

echo "boundary check passed"
