#!/bin/bash
set -e
# 首次服务器初始化脚本 (Ubuntu/Debian)

echo "🔧 ChillCat 服务器初始化"
echo "========================"

# 安装 Docker
if ! command -v docker &> /dev/null; then
    echo "📦 安装 Docker..."
    curl -fsSL https://get.docker.com | sh
    sudo usermod -aG docker $USER
fi

# 安装 Nginx
if ! command -v nginx &> /dev/null; then
    echo "📦 安装 Nginx..."
    sudo apt update && sudo apt install -y nginx certbot python3-certbot-nginx
fi

# 配置 Nginx
echo "🔧 配置 Nginx..."
sudo cp "$(dirname "$0")/nginx.conf" /etc/nginx/sites-available/chillcat
sudo ln -sf /etc/nginx/sites-available/chillcat /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx

# SSL 证书
echo ""
read -p "是否申请 Let's Encrypt SSL 证书？(y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    sudo certbot --nginx -d api.chillcatgo.com
fi

echo ""
echo "✅ 服务器初始化完成！"
echo "   下一步: cd /path/to/server && bash deploy/deploy.sh"
