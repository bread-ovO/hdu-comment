#!/bin/bash

# 邮箱验证功能测试脚本

set -e

echo "=== 邮箱验证功能测试 ==="

# 1. 检查环境变量
echo "1. 检查环境配置..."
if [ -z "$SMTP_USERNAME" ] || [ -z "$SMTP_PASSWORD" ]; then
    echo "警告: SMTP_USERNAME 或 SMTP_PASSWORD 未设置，将使用模拟模式"
    echo "请设置以下环境变量："
    echo "  SMTP_HOST - SMTP服务器地址 (默认: smtp.gmail.com)"
    echo "  SMTP_PORT - SMTP端口 (默认: 587)"
    echo "  SMTP_USERNAME - SMTP用户名"
    echo "  SMTP_PASSWORD - SMTP密码"
    echo "  FROM_EMAIL - 发件人邮箱"
fi

# 2. 运行数据库迁移
echo "2. 运行数据库迁移..."
cd backend
go run cmd/migrate/main.go

# 3. 启动后端服务
echo "3. 启动后端服务..."
go run cmd/server/main.go &

# 等待服务启动
sleep 5

# 4. 测试API端点
echo "4. 测试API端点..."

# 请求注册验证码
echo "测试发送注册验证码接口..."
curl -X POST http://localhost:8080/api/v1/auth/send-code \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com"
  }'

echo ""
echo "请从邮箱中获取验证码后，替换下面命令中的 <CODE> 再执行："
echo ""
echo "curl -X POST http://localhost:8080/api/v1/auth/register \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -d '{"
echo "    \"email\": \"test@example.com\","
echo "    \"password\": \"test123456\","
echo "    \"display_name\": \"测试用户\","
echo "    \"code\": \"<CODE>\""
echo "  }'"

# 测试邮箱验证
echo "测试邮箱验证接口..."
# 这里需要实际的验证令牌，暂时跳过

# 5. 测试前端页面
echo "5. 测试前端页面..."
echo "请访问 http://localhost:5174/register 进行注册测试"
echo "请访问 http://localhost:5174/verify-email?token=YOUR_TOKEN 进行验证测试"

echo "=== 测试完成 ==="
