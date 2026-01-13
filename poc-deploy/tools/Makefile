.PHONY: help generate generate-shell submit quick-test clean build

# 默认参数
NUM_ACCOUNTS ?= 2
FILES_PER_ACCOUNT ?= 50
ORDERS_PER_FILE ?= 200

help:
	@echo "批量订单测试工具"
	@echo ""
	@echo "用法:"
	@echo "  make generate          - 生成测试订单文件（Go 版本，推荐）"
	@echo "  make generate-shell    - 生成测试订单文件（Shell 版本）"
	@echo "  make submit            - 提交订单到链上"
	@echo "  make quick-test        - 快速测试（生成 + 提交）"
	@echo "  make build             - 编译 Go 程序"
	@echo "  make clean             - 清理生成的文件"
	@echo ""
	@echo "参数:"
	@echo "  NUM_ACCOUNTS=$(NUM_ACCOUNTS)           - 账户数量"
	@echo "  FILES_PER_ACCOUNT=$(FILES_PER_ACCOUNT)   - 每账户文件数"
	@echo "  ORDERS_PER_FILE=$(ORDERS_PER_FILE)   - 每文件订单数"
	@echo ""
	@echo "示例:"
	@echo "  make generate NUM_ACCOUNTS=3 FILES_PER_ACCOUNT=100 ORDERS_PER_FILE=500"
	@echo "  make submit NUM_ACCOUNTS=3"
	@echo "  make quick-test NUM_ACCOUNTS=2 FILES_PER_ACCOUNT=10 ORDERS_PER_FILE=100"
	@echo ""

generate:
	@echo "生成测试订单（Go 版本）..."
	go run generate-orders.go $(NUM_ACCOUNTS) $(FILES_PER_ACCOUNT) $(ORDERS_PER_FILE)

generate-shell:
	@echo "生成测试订单（Shell 版本）..."
	./generate-test-orders.sh $(NUM_ACCOUNTS) $(FILES_PER_ACCOUNT) $(ORDERS_PER_FILE)

submit:
	@echo "提交订单..."
	go run batch-submit.go --count $(NUM_ACCOUNTS)

quick-test:
	@echo "快速测试..."
	./quick-test.sh $(NUM_ACCOUNTS) $(FILES_PER_ACCOUNT) $(ORDERS_PER_FILE)

build:
	@echo "编译 Go 程序..."
	go build -o batch-submit batch-submit.go
	go build -o generate-orders generate-orders.go
	@echo "✓ 编译完成: ./batch-submit, ./generate-orders"

clean:
	@echo "清理生成的文件..."
	rm -rf order*/ batch-submit generate-orders
	@echo "✓ 清理完成"

