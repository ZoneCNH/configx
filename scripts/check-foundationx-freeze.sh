#!/usr/bin/env bash
# check-foundationx-freeze.sh — 强制执行 configx foundationx 冻结策略
#
# 策略（FOUNDATION-DEPS.yaml §no-foundationx-new-usage）：
#   冻结 foundationx 用法：允许已有引用继续存在，禁止新增引用。
#   迁移路径：docs/foundationx-compatibility.md → v0.3 前完成 → 删除 internal/foundationx/
#
# 机制：基线文件（scripts/.foundationx-baseline.txt）记录当前允许的引用位置。
#       任何新增引用（不在基线中的文件）都会触发失败。
#       当引用被移除时，运行 --update-baseline 更新基线。

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BASELINE="$SCRIPT_DIR/.foundationx-baseline.txt"

cd "$REPO_ROOT"

# 生成当前 foundationx 引用清单（排除 internal/foundationx 自身和 worktree 目录）
generate_manifest() {
    grep -rn "foundationx" --include="*.go" . \
        | grep -v "_test.go" \
        | grep -v "internal/foundationx" \
        | grep -v "/\.worktree/" \
        | cut -d: -f1 \
        | sort -u
}

if [ "${1:-}" = "--update-baseline" ]; then
    generate_manifest > "$BASELINE"
    echo "✅ foundationx baseline updated ($(wc -l < "$BASELINE") files)"
    exit 0
fi

if [ ! -f "$BASELINE" ]; then
    echo "⚠️  No baseline found. Creating initial baseline..."
    generate_manifest > "$BASELINE"
    echo "✅ Initial baseline created ($(wc -l < "$BASELINE") files)"
    echo "   Run with --update-baseline to refresh after removing references."
    exit 0
fi

CURRENT=$(generate_manifest)
BASELINE_CONTENT=$(cat "$BASELINE")

# 找出新增文件（不在基线中）
NEW_FILES=$(comm -13 <(echo "$BASELINE_CONTENT" | sort) <(echo "$CURRENT" | sort))

if [ -n "$NEW_FILES" ]; then
    echo "❌ FOUNDATIONX FREEZE VIOLATION — 检测到新增 foundationx 引用"
    echo ""
    echo "新增文件："
    echo "$NEW_FILES" | while read -r f; do
        echo "  $f"
        grep -n "foundationx" "$f" | head -3 | while read -r line; do
            echo "    $line"
        done
    done
    echo ""
    echo "策略：不再新增 foundationx usage（FOUNDATION-DEPS.yaml §no-foundationx-new-usage）"
    echo "迁移：docs/foundationx-compatibility.md → v0.3 → 删除 internal/foundationx/"
    echo "如需接受新引用（仅限迁移过渡期），运行："
    echo "  scripts/check-foundationx-freeze.sh --update-baseline"
    exit 1
fi

# 检查是否有文件被移除（迁移进展）
REMOVED=$(comm -23 <(echo "$BASELINE_CONTENT" | sort) <(echo "$CURRENT" | sort))
if [ -n "$REMOVED" ]; then
    echo "📉 foundationx 引用减少 — 迁移进展"
    echo "$REMOVED" | while read -r f; do echo "  ✅ 已移除: $f"; done
    echo "  运行 --update-baseline 更新基线"
fi

echo "✅ foundationx freeze check passed — $(echo "$CURRENT" | wc -l) 个文件，无新增引用"
