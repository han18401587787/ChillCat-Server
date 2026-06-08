#!/bin/bash
set -e
# ChillCat 服务器初始化脚本 (支持 Debian/Ubuntu 和 RHEL/CentOS/OpenCloudOS)

echo "🔧 ChillCat 服务器初始化"
echo "========================"

# 检测包管理器
if command -v apt &> /dev/null; then
    PKG_MGR="apt"
elif command -v dnf &> /dev/null; then
    PKG_MGR="dnf"
elif command -v yum &> /dev/null; then
    PKG_MGR="yum"
else
    echo "❌ 无法检测包管理器 (apt/dnf/yum)"
    exit 1
fi
echo "📦 包管理器: $PKG_MGR"

# 安装 Docker
if ! command -v docker &> /dev/null; then
    echo "📦 安装 Docker..."
    curl -fsSL https://get.docker.com | sh
    sudo systemctl enable docker
    sudo systemctl start docker
    sudo usermod -aG docker $USER
    echo "⚠️ 请重新登录以使 Docker 权限生效"
fi

# 安装 Docker Compose
if ! command -v docker &> /dev/null; then
    :
elif ! docker compose version &> /dev/null 2>&1; then
    echo "📦 安装 Docker Compose..."
    sudo $PKG_MGR install -y docker-compose-plugin 2>/dev/null || true
fi

# 安装 Nginx
if ! command -v nginx &> /dev/null; then
    echo "📦 安装 Nginx..."
    case $PKG_MGR in
        apt)
            sudo apt update && sudo apt install -y nginx certbot python3-certbot-nginx
            NGINX_CONF_DIR="/etc/nginx/sites-available"
            NGINX_ENABLED_DIR="/etc/nginx/sites-enabled"
            ;;
        dnf|yum)
            sudo $PKG_MGR install -y nginx certbot python3-certbot-nginx 2>/dev/null || \
            sudo $PKG_MGR install -y nginx epel-release && sudo $PKG_MGR install -y certbot python3-certbot-nginx
            NGINX_CONF_DIR="/etc/nginx/conf.d"
            NGINX_ENABLED_DIR="/etc/nginx/conf.d"
            ;;
    esac
else
    # Nginx 已安装，确定配置目录
    if [ -d "/etc/nginx/sites-available" ]; then
        NGINX_CONF_DIR="/etc/nginx/sites-available"
        NGINX_ENABLED_DIR="/etc/nginx/sites-enabled"
    else
        NGINX_CONF_DIR="/etc/nginx/conf.d"
        NGINX_ENABLED_DIR="/etc/nginx/conf.d"
    fi
fi

echo "📁 Nginx 配置目录: $NGINX_CONF_DIR"

# 配置 Nginx
echo "🔧 配置 Nginx..."
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CONF_FILE="$NGINX_CONF_DIR/chillcat.conf"
sudo cp "$SCRIPT_DIR/nginx.conf" "$CONF_FILE"

# RHEL 系不需要 sites-enabled 符号链接
if [ "$NGINX_CONF_DIR" != "$NGINX_ENABLED_DIR" ]; then
    sudo ln -sf "$CONF_FILE" "$NGINX_ENABLED_DIR/chillcat.conf"
fi

# 测试并重载 Nginx
sudo nginx -t && sudo systemctl reload nginx 2>/dev/null || sudo systemctl start nginx
echo "✅ Nginx 配置完成"

# 开放防火墙端口
if command -v firewall-cmd &> /dev/null; then
    sudo firewall-cmd --permanent --add-service=http 2>/dev/null || true
    sudo firewall-cmd --permanent --add-service=https 2>/dev/null || true
    sudo firewall-cmd --reload 2>/dev/null || true
fi

# SSL 证书
echo ""
read -p "是否申请 Let's Encrypt SSL 证书？(y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    sudo certbot --nginx -d api.chillcatgo.com
fi

echo ""
echo "✅ 服务器初始化完成！"
echo "   下一步: bash deploy/deploy.sh"
