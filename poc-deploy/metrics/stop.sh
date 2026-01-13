#!/usr/bin/env bash

set -e

echo "=========================================="
echo "停止 Sei Chain 监控服务"
echo "=========================================="

# 停止服务
docker-compose down

echo ""
echo "监控服务已停止"
echo ""
echo "如需删除数据卷，运行："
echo "  docker-compose down -v"
echo ""

