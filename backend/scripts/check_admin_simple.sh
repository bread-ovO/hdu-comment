#!/bin/bash

echo "=== 检查管理员账户 ==="
echo ""

# 检查数据库文件
if [ -f "data/app.db" ]; then
    echo "✅ 数据库文件存在: data/app.db"
    
    # 检查是否有sqlite3命令
    if command -v sqlite3 &> /dev/null; then
        echo ""
        echo "=== 所有用户列表 ==="
        sqlite3 data/app.db "SELECT id, email, display_name, role, created_at FROM users ORDER BY created_at;"
        
        echo ""
        echo "=== 管理员用户统计 ==="
        ADMIN_COUNT=$(sqlite3 data/app.db "SELECT COUNT(*) FROM users WHERE role='admin';")
        
        if [ "$ADMIN_COUNT" -eq 0 ]; then
            echo "❌ 没有找到管理员账户 (role='admin')"
            echo ""
            echo "=== 解决方案 ==="
            echo "1. 设置环境变量:"
            echo "   export APP_ADMIN_EMAIL=admin@yourdomain.com"
            echo "   export APP_ADMIN_PASSWORD=your-password"
            echo "   export APP_AUTH_JWT_SECRET=your-secret-key"
            echo "2. 重启应用: docker-compose restart backend"
            echo "3. 应用会自动创建管理员账户"
        else
            echo "✅ 找到 $ADMIN_COUNT 个管理员账户:"
            sqlite3 data/app.db "SELECT id, email, display_name, created_at FROM users WHERE role='admin';"
        fi
    else
        echo "❌ sqlite3 未安装，无法直接查询数据库"
        echo "请安装 sqlite3: apt-get install sqlite3 或 brew install sqlite3"
    fi
else
    echo "❌ 数据库文件不存在: data/app.db"
    echo "请确保应用已启动并初始化数据库"
fi

echo ""
echo "=== 环境变量检查 ==="
echo "APP_ADMIN_EMAIL: ${APP_ADMIN_EMAIL:-未设置}"
echo "APP_ADMIN_PASSWORD: ${APP_ADMIN_PASSWORD:-未设置}"
echo "APP_AUTH_JWT_SECRET: ${APP_AUTH_JWT_SECRET:-未设置}"