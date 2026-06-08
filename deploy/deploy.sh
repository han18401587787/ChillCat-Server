#!/bin/bash
set -e

echo "🚀 ChillCat 服务端部署脚本"
echo "============================"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR"

# 拉取最新代码
echo ""
echo "📦 拉取最新代码..."
git pull origin main

# 构建镜像
echo ""
echo "🔨 构建 Docker 镜像..."
docker compose -f docker-compose.prod.yml build --no-cache

# 重启服务
echo ""
echo "🔄 重启服务..."
docker compose -f docker-compose.prod.yml down
docker compose -f docker-compose.prod.yml up -d

# 等待健康检查
echo ""
echo "⏳ 等待服务就绪..."
sleep 5

# 健康检查
echo ""
echo "🏥 健康检查..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health || echo "000")
if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ 服务运行正常 (HTTP $HTTP_CODE)"
else
    echo "❌ 健康检查失败 (HTTP $HTTP_CODE)"
    docker compose -f docker-compose.prod.yml logs --tail=20
    exit 1
fi

# 填充种子数据（仅首次）
echo ""
read -p "是否填充种子数据？(y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    docker compose -f docker-compose.prod.yml exec -T app ./server --seed 2>/dev/null || echo "⚠️ 种子数据填充需在容器内手动执行"
fi

echo ""
echo "🎉 部署完成！"
echo "   API: http://localhost:8080"
echo "   查看日志: docker compose -f docker-compose.prod.yml logs -f"
