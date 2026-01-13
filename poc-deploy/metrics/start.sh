#!/usr/bin/env bash

set -e

echo "=========================================="
echo "启动 Sei Chain 监控服务"
echo "=========================================="

# 检查 Docker 是否运行
if ! docker info > /dev/null 2>&1; then
    echo "错误: Docker 未运行，请先启动 Docker"
    exit 1
fi

# 检查 docker-compose 是否安装
if ! command -v docker-compose &> /dev/null; then
    echo "错误: docker-compose 未安装"
    echo "请安装 docker-compose: https://docs.docker.com/compose/install/"
    exit 1
fi

# 启动服务
echo ""
echo "启动 Prometheus 和 Grafana..."
docker-compose up -d

# 等待服务启动
echo ""
echo "等待服务启动..."
sleep 5

# 检查服务状态
echo ""
echo "检查服务状态..."
docker-compose ps

# 显示访问信息
echo ""
echo "=========================================="
echo "监控服务已启动！"
echo "=========================================="
echo ""
echo "访问地址："
echo "  - Grafana:    http://localhost:3000"
echo "    用户名: admin"
echo "    密码: admin"
echo ""
echo "  - Prometheus: http://localhost:9090"
echo ""
echo "查看日志："
echo "  docker-compose logs -f"
echo ""
echo "停止服务："
echo "  docker-compose down"
echo ""

