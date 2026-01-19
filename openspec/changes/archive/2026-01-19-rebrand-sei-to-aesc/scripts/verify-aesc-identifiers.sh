#!/bin/bash
# 验证 AESC 标识符的正确配置

set -e

echo "=== 验证 AESC 标识符配置 ==="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 错误计数
ERRORS=0

echo "1. 验证 app/params/config.go："
echo "----------------------------------------------------------------"
if [ -f "app/params/config.go" ]; then
    CHECKS=0
    PASSED=0
    
    # 检查 BaseCoinUnit
    CHECKS=$((CHECKS + 1))
    if grep -q 'BaseCoinUnit.*=.*"uaex"' app/params/config.go; then
        echo -e "${GREEN}✓ BaseCoinUnit = \"uaex\"${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ BaseCoinUnit 配置错误${NC}"
        ERRORS=$((ERRORS + 1))
    fi
    
    # 检查 HumanCoinUnit
    CHECKS=$((CHECKS + 1))
    if grep -q 'HumanCoinUnit.*=.*"aex"' app/params/config.go; then
        echo -e "${GREEN}✓ HumanCoinUnit = \"aex\"${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ HumanCoinUnit 配置错误${NC}"
        ERRORS=$((ERRORS + 1))
    fi
    
    # 检查 Bech32PrefixAccAddr
    CHECKS=$((CHECKS + 1))
    if grep -q 'Bech32PrefixAccAddr.*=.*"aesc"' app/params/config.go; then
        echo -e "${GREEN}✓ Bech32PrefixAccAddr = \"aesc\"${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ Bech32PrefixAccAddr 配置错误${NC}"
        ERRORS=$((ERRORS + 1))
    fi
    
    # 检查 UaexExponent
    CHECKS=$((CHECKS + 1))
    if grep -q 'UaexExponent.*=.*6' app/params/config.go; then
        echo -e "${GREEN}✓ UaexExponent = 6${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${YELLOW}⚠ UaexExponent 可能未配置（可选）${NC}"
    fi
    
    echo "app/params/config.go: $PASSED/$CHECKS 检查通过"
else
    echo -e "${RED}✗ app/params/config.go 不存在${NC}"
    ERRORS=$((ERRORS + 1))
fi
echo ""

echo "2. 验证 EVM 模块配置："
echo "----------------------------------------------------------------"
EVM_FOUND=false

# 检查可能的位置
for file in "x/evm/types/params.go" "x/evm/keeper/params.go"; do
    if [ -f "$file" ]; then
        if grep -q 'BaseDenom.*=.*"uaex"' "$file"; then
            echo -e "${GREEN}✓ $file: BaseDenom = \"uaex\"${NC}"
            EVM_FOUND=true
            break
        fi
    fi
done

if [ "$EVM_FOUND" = false ]; then
    echo -e "${RED}✗ EVM 模块 BaseDenom 配置未找到或错误${NC}"
    echo "  检查的文件："
    echo "  - x/evm/types/params.go"
    echo "  - x/evm/keeper/params.go"
    ERRORS=$((ERRORS + 1))
fi
echo ""

echo "3. 验证 Makefile 测试命令："
echo "----------------------------------------------------------------"
if [ -f "Makefile" ]; then
    CHECKS=0
    PASSED=0
    
    # 检查 test 命令
    CHECKS=$((CHECKS + 1))
    if grep -q "^test:" Makefile; then
        echo -e "${GREEN}✓ 'make test' 命令存在${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ 'make test' 命令不存在${NC}"
        ERRORS=$((ERRORS + 1))
    fi
    
    # 检查 test-unit 命令
    CHECKS=$((CHECKS + 1))
    if grep -q "^test-unit:" Makefile; then
        echo -e "${GREEN}✓ 'make test-unit' 命令存在${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${YELLOW}⚠ 'make test-unit' 命令不存在（建议添加）${NC}"
    fi
    
    # 检查 test-integration 命令
    CHECKS=$((CHECKS + 1))
    if grep -q "^test-integration:" Makefile; then
        echo -e "${GREEN}✓ 'make test-integration' 命令存在${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${YELLOW}⚠ 'make test-integration' 命令不存在（建议添加）${NC}"
    fi
    
    echo "Makefile: $PASSED/$CHECKS 检查通过"
else
    echo -e "${RED}✗ Makefile 不存在${NC}"
    ERRORS=$((ERRORS + 1))
fi
echo ""

echo "4. 验证 cmd/seid 配置："
echo "----------------------------------------------------------------"
if [ -f "cmd/seid/cmd/root.go" ]; then
    if grep -q 'MinGasPrices.*=.*".*uaex"' cmd/seid/cmd/root.go; then
        echo -e "${GREEN}✓ MinGasPrices 使用 uaex${NC}"
    else
        echo -e "${YELLOW}⚠ MinGasPrices 可能未配置或使用其他面额${NC}"
        echo "  请手动检查 cmd/seid/cmd/root.go"
    fi
else
    echo -e "${YELLOW}⚠ cmd/seid/cmd/root.go 不存在${NC}"
fi
echo ""

echo "5. 检查示例 Genesis 配置："
echo "----------------------------------------------------------------"
GENESIS_FILES=$(find . -name "*genesis*.json" ! -path "*/vendor/*" ! -path "*/.git/*" 2>/dev/null)

if [ -n "$GENESIS_FILES" ]; then
    echo "找到以下 genesis 文件："
    echo "$GENESIS_FILES"
    echo ""
    
    for file in $GENESIS_FILES; do
        echo "检查: $file"
        
        # 检查是否包含 uaex
        if grep -q "uaex" "$file"; then
            echo -e "${GREEN}  ✓ 包含 uaex 面额${NC}"
        else
            echo -e "${YELLOW}  ⚠ 未找到 uaex 面额${NC}"
        fi
        
        # 检查是否包含 aesc1 地址
        if grep -q "aesc1" "$file"; then
            echo -e "${GREEN}  ✓ 包含 aesc1 地址${NC}"
        else
            echo -e "${YELLOW}  ⚠ 未找到 aesc1 地址${NC}"
        fi
        
        # 检查是否仍包含 usei（不应该有）
        if grep -q "usei" "$file"; then
            echo -e "${RED}  ✗ 仍包含 usei 面额（需要修改）${NC}"
            ERRORS=$((ERRORS + 1))
        fi
        
        # 检查是否仍包含 sei1 地址（不应该有）
        if grep -q "sei1" "$file"; then
            echo -e "${RED}  ✗ 仍包含 sei1 地址（需要修改）${NC}"
            ERRORS=$((ERRORS + 1))
        fi
        echo ""
    done
else
    echo -e "${YELLOW}⚠ 未找到 genesis 配置文件${NC}"
fi
echo ""

echo "=== 验证完成 ==="
echo ""

if [ "$ERRORS" -eq 0 ]; then
    echo -e "${GREEN}✓ 所有关键配置验证通过${NC}"
    echo ""
    echo "建议："
    echo "1. 运行 'make test' 验证功能"
    echo "2. 启动本地节点测试配置"
    echo "3. 检查所有测试文件是否使用正确的标识符"
    exit 0
else
    echo -e "${RED}✗ 发现 $ERRORS 个配置错误${NC}"
    echo ""
    echo "请修复上述错误后重新运行验证"
    exit 1
fi

