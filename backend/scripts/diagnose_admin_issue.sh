#!/bin/bash

# 管理员权限问题诊断脚本

echo "=== 管理员删除帖子权限问题诊断 ==="
echo ""

# 检查环境变量
echo "=== 环境变量检查 ==="
echo "APP_ADMIN_EMAIL: ${APP_ADMIN_EMAIL:-未设置}"
echo "APP_ADMIN_PASSWORD: ${APP_ADMIN_PASSWORD:-未设置}"
echo "APP_AUTH_JWT_SECRET: ${APP_AUTH_JWT_SECRET:-未设置}"
echo ""

# 检查数据库文件
echo "=== 数据库检查 ==="
if [ -f "data/app.db" ]; then
    echo "数据库文件存在: data/app.db"
    
    # 使用sqlite3检查管理员用户（如果可用）
    if command -v sqlite3 &> /dev/null; then
        echo "管理员用户信息:"
        sqlite3 data/app.db "SELECT id, email, display_name, role FROM users WHERE email='${APP_ADMIN_EMAIL:-admin@example.com}';"
    else
        echo "sqlite3 未安装，无法直接查询数据库"
    fi
else
    echo "数据库文件不存在: data/app.db"
fi

echo ""
echo "=== 建议的修复步骤 ==="
echo "1. 确保服务器上设置了正确的环境变量："
echo "   export APP_ADMIN_EMAIL=your-admin@email.com"
echo "   export APP_ADMIN_PASSWORD=your-admin-password"
echo "   export APP_AUTH_JWT_SECRET=your-jwt-secret"
echo ""
echo "2. 重启应用以重新初始化管理员用户"
echo ""
echo "3. 使用管理员账号重新登录获取新的JWT令牌"