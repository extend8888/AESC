#!/bin/bash
# 检查代码库中残留的 Sei 标识符

set -e

echo "=== 检查代码库中的 Sei 标识符 ==="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 计数器
ISSUES_FOUND=0

echo "1. 检查 Go 文件中的 usei 引用（排除注释和第三方代码）："
echo "----------------------------------------------------------------"
USEI_FILES=$(find . -type f -name "*.go" \
  ! -path "*/vendor/*" \
  ! -path "*/go-ethereum/*" \
  ! -path "*/.git/*" \
  -exec grep -l "usei" {} \; 2>/dev/null | wc -l)

if [ "$USEI_FILES" -gt 0 ]; then
    echo -e "${YELLOW}发现 $USEI_FILES 个文件包含 'usei'${NC}"
    find . -type f -name "*.go" \
      ! -path "*/vendor/*" \
      ! -path "*/go-ethereum/*" \
      ! -path "*/.git/*" \
      -exec grep -Hn "usei" {} \; 2>/dev/null | head -20
    echo "（显示前 20 个结果）"
    ISSUES_FOUND=$((ISSUES_FOUND + USEI_FILES))
else
    echo -e "${GREEN}✓ 未发现 usei 引用${NC}"
fi
echo ""

echo "2. 检查 Go 文件中的 sei1 地址："
echo "----------------------------------------------------------------"
SEI1_FILES=$(find . -type f -name "*.go" \
  ! -path "*/vendor/*" \
  ! -path "*/go-ethereum/*" \
  ! -path "*/.git/*" \
  -exec grep -l "sei1" {} \; 2>/dev/null | wc -l)

if [ "$SEI1_FILES" -gt 0 ]; then
    echo -e "${YELLOW}发现 $SEI1_FILES 个文件包含 'sei1' 地址${NC}"
    find . -type f -name "*.go" \
      ! -path "*/vendor/*" \
      ! -path "*/go-ethereum/*" \
      ! -path "*/.git/*" \
      -exec grep -Hn "sei1" {} \; 2>/dev/null | head -20
    echo "（显示前 20 个结果）"
    ISSUES_FOUND=$((ISSUES_FOUND + SEI1_FILES))
else
    echo -e "${GREEN}✓ 未发现 sei1 地址${NC}"
fi
echo ""

echo "3. 检查配置文件中的 usei："
echo "----------------------------------------------------------------"
CONFIG_FILES=$(find . -type f \( -name "*.json" -o -name "*.toml" -o -name "*.yaml" -o -name "*.yml" \) \
  ! -path "*/vendor/*" \
  ! -path "*/node_modules/*" \
  ! -path "*/.git/*" \
  -exec grep -l "usei" {} \; 2>/dev/null | wc -l)

if [ "$CONFIG_FILES" -gt 0 ]; then
    echo -e "${YELLOW}发现 $CONFIG_FILES 个配置文件包含 'usei'${NC}"
    find . -type f \( -name "*.json" -o -name "*.toml" -o -name "*.yaml" -o -name "*.yml" \) \
      ! -path "*/vendor/*" \
      ! -path "*/node_modules/*" \
      ! -path "*/.git/*" \
      -exec grep -Hn "usei" {} \; 2>/dev/null | head -20
    echo "（显示前 20 个结果）"
    ISSUES_FOUND=$((ISSUES_FOUND + CONFIG_FILES))
else
    echo -e "${GREEN}✓ 未发现配置文件中的 usei${NC}"
fi
echo ""

echo "4. 检查脚本文件中的 usei："
echo "----------------------------------------------------------------"
SCRIPT_FILES=$(find . -type f -name "*.sh" \
  ! -path "*/vendor/*" \
  ! -path "*/.git/*" \
  -exec grep -l "usei" {} \; 2>/dev/null | wc -l)

if [ "$SCRIPT_FILES" -gt 0 ]; then
    echo -e "${YELLOW}发现 $SCRIPT_FILES 个脚本文件包含 'usei'${NC}"
    find . -type f -name "*.sh" \
      ! -path "*/vendor/*" \
      ! -path "*/.git/*" \
      -exec grep -Hn "usei" {} \; 2>/dev/null | head -20
    echo "（显示前 20 个结果）"
    ISSUES_FOUND=$((ISSUES_FOUND + SCRIPT_FILES))
else
    echo -e "${GREEN}✓ 未发现脚本文件中的 usei${NC}"
fi
echo ""

echo "5. 检查 Makefile 中的 sei 引用（排除 seid）："
echo "----------------------------------------------------------------"
if [ -f "Makefile" ]; then
    MAKEFILE_ISSUES=$(grep -n "sei" Makefile | grep -v "seid" | grep -v "^#" | wc -l)
    if [ "$MAKEFILE_ISSUES" -gt 0 ]; then
        echo -e "${YELLOW}发现 $MAKEFILE_ISSUES 个 Makefile 中的 sei 引用${NC}"
        grep -n "sei" Makefile | grep -v "seid" | grep -v "^#"
        ISSUES_FOUND=$((ISSUES_FOUND + MAKEFILE_ISSUES))
    else
        echo -e "${GREEN}✓ Makefile 中未发现问题${NC}"
    fi
else
    echo -e "${YELLOW}⚠ Makefile 不存在${NC}"
fi
echo ""

echo "=== 检查完成 ==="
echo ""

if [ "$ISSUES_FOUND" -eq 0 ]; then
    echo -e "${GREEN}✓ 所有检查通过，未发现残留的 Sei 标识符${NC}"
    exit 0
else
    echo -e "${RED}✗ 发现 $ISSUES_FOUND 个潜在问题，请检查上述输出${NC}"
    echo ""
    echo "注意：某些引用可能是合理的（如注释、文档、第三方代码等）"
    echo "请手动审查这些引用，确定是否需要修改"
    exit 1
fi

