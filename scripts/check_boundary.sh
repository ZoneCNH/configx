#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "checking forbidden dependency boundary..."

DEPS="$(GOWORK="${GOWORK:-off}" go list -deps ./...)"
FORBIDDEN_DEPS=(
  "github.com/bytechainx/x.go"
  "github.com/ZoneCNH/x.go"
  "database/sql"
  "github.com/jackc/pgx"
  "github.com/lib/pq"
  "github.com/go-sql-driver/mysql"
  "github.com/segmentio/kafka-go"
  "github.com/IBM/sarama"
  "github.com/Shopify/sarama"
  "github.com/confluentinc/confluent-kafka-go"
  "github.com/redis/go-redis"
  "github.com/taosdata"
  "github.com/aws/aws-sdk-go"
  "github.com/aws/aws-sdk-go-v2"
  "github.com/aliyun"
  "github.com/minio/minio-go"
)

for dep in "${FORBIDDEN_DEPS[@]}"; do
  if grep -Fq "$dep" <<<"$DEPS"; then
    echo "ERROR: forbidden infrastructure dependency found: $dep"
    exit 1
  fi
done

echo "checking forbidden implicit config discovery..."

SEARCH_DIRS=()
for dir in pkg internal examples contracts; do
  if [ -d "$dir" ]; then
    SEARCH_DIRS+=("$dir")
  fi
done

IMPLICIT_DISCOVERY_PATTERNS=(
  "godotenv"
  "AutomaticEnv"
  "SetConfigName"
  "AddConfigPath"
  "FindConfig"
  "ReadInConfig"
  "UserConfigDir"
  "UserHomeDir"
)

for pattern in "${IMPLICIT_DISCOVERY_PATTERNS[@]}"; do
  if [ "${#SEARCH_DIRS[@]}" -gt 0 ] && grep -R --line-number --fixed-strings "$pattern" "${SEARCH_DIRS[@]}" --exclude-dir=.git; then
    echo "ERROR: implicit config discovery pattern found: $pattern"
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
