#!/usr/bin/env bash

SERVICE=${1:-}

if [ -z "$SERVICE" ]; then
    echo "查看所有服务日志..."
    docker-compose logs -f
else
    echo "查看 $SERVICE 服务日志..."
    docker-compose logs -f "$SERVICE"
fi

