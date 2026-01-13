#!/usr/bin/env bash

# 模拟 batch-submit.go 生成的命令
echo "生成的命令格式："
echo "echo 12345678 | seid tx execution batch-ingest <file> --from admin1 --chain-id sei-poc --node tcp://localhost:26657 --fees 2000000usei --gas 50000000 --broadcast-mode sync -y"
echo ""
echo "你的命令格式："
echo "echo 12345678 | seid tx execution batch-ingest orders-0001.json --from admin1 -y --fees 2000000usei --gas 50000000"
echo ""
echo "主要区别："
echo "1. 添加了 --chain-id sei-poc"
echo "2. 添加了 --node tcp://localhost:26657"
echo "3. 添加了 --broadcast-mode sync"
echo ""
echo "这些参数都是必需的，命令应该是正确的。"
