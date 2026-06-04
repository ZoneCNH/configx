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
  "github.com/redis/rueidis"
  "github.com/IBM/sarama"
  "github.com/Shopify/sarama"
  "github.com/segmentio/kafka-go"
  "github.com/confluentinc/confluent-kafka-go"
  "github.com/jackc/pgx"
  "github.com/lib/pq"
  "gorm.io/gorm"
  "go.mongodb.org/mongo-driver"
  "github.com/taosdata/driver-go"
  "github.com/aws/aws-sdk-go"
  "github.com/aws/aws-sdk-go-v2"
  "github.com/aws/aws-sdk-go/service/kms"
  "github.com/aws/aws-sdk-go-v2/service/kms"
  "cloud.google.com/go"
  "google.golang.org/api"
  "github.com/Azure/azure-sdk-for-go"
  "github.com/Azure/azure-sdk-for-go/sdk"
  "github.com/hashicorp/vault"
  "github.com/hashicorp/vault/api"
  "github.com/hashicorp/consul"
  "github.com/hashicorp/consul/api"
  "go.etcd.io/etcd"
  "github.com/nacos-group/nacos-sdk-go"
  "github.com/aliyun/aliyun-oss-go-sdk"
  "github.com/aliyun/alibaba-cloud-sdk-go"
  "github.com/minio/minio-go"
)

for dep in "${FORBIDDEN_DEPS[@]}"; do
  if grep -Fq "$dep" <<<"$DEPS"; then
    echo "ERROR: base library template must not depend on forbidden infrastructure dependency: $dep"
    exit 1
  fi
done

SEARCH_DIRS=(pkg internal contracts examples scripts release testkit Makefile)
GREP_EXCLUDES=(
  --exclude-dir=.git
  --exclude=check_boundary.sh
  --exclude=check_release_evidence.sh
)

echo "checking forbidden implicit config discovery..."

FORBIDDEN_DISCOVERY_PATTERNS=(
  '(^|[^[:alnum:]_./-])\.env([^[:alnum:]_./-]|$)'
  '(^|[^[:alnum:]_./-])production\.yaml([^[:alnum:]_./-]|$)'
  '(^|[^[:alnum:]_./-])production\.yml([^[:alnum:]_./-]|$)'
  '(^|[^[:alnum:]_./-])config\.local\.yaml([^[:alnum:]_./-]|$)'
  '(^|[^[:alnum:]_./-])config\.local\.yml([^[:alnum:]_./-]|$)'
  '/home/k8s/secrets/env'
)

for pattern in "${FORBIDDEN_DISCOVERY_PATTERNS[@]}"; do
  if grep -R --line-number --extended-regexp "$pattern" "${SEARCH_DIRS[@]}" "${GREP_EXCLUDES[@]}"; then
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
  if [ "${#SEARCH_DIRS[@]}" -gt 0 ] && grep -R --line-number --fixed-strings "$term" "${SEARCH_DIRS[@]}" "${GREP_EXCLUDES[@]}"; then
    echo "ERROR: forbidden business term found: $term"
    exit 1
  fi
done

echo "boundary check passed"
