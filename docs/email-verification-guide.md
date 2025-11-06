# 邮箱验证功能使用指南

## 功能概述
邮箱验证功能确保用户注册时提供有效的邮箱地址，提高账户安全性和用户体验。

## 系统架构

### 后端组件
- **EmailVerification模型**: 存储验证令牌信息
- **EmailVerificationRepository**: 数据访问层
- **EmailVerificationService**: 业务逻辑层
- **EmailVerificationHandler**: API接口层
- **EmailService**: 邮件发送服务

### 前端组件
- **VerifyEmail页面**: 邮箱验证页面
- **Register页面**: 注册页面集成邮箱验证
- **EmailVerificationAlert**: 邮箱验证提醒组件
- **Auth API**: 认证相关API调用

## 配置说明

### 环境变量
```bash
# 邮件服务配置
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
FROM_EMAIL=noreply@yourdomain.com
FROM_NAME=HDU美食点评
```

### 数据库迁移
```bash
# 运行数据库迁移
cd backend
go run cmd/migrate/main.go
```

## API接口

### 1. 发送注册验证码
```http
POST /api/v1/auth/send-code
Content-Type: application/json

{
  "email": "user@example.com"
}
```

### 2. 提交注册
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "******",
  "display_name": "昵称",
  "code": "123456"
}
```

### 3. 发送验证邮件（登录状态）
```http
POST /api/v1/auth/send-verification
Authorization: Bearer <token>
```

### 4. 验证邮箱
```http
POST /api/v1/auth/verify-email
Content-Type: application/json

{
  "token": "verification-token"
}
```

### 5. 获取验证状态
```http
GET /api/v1/auth/verification-status
Authorization: Bearer <token>
```

## 使用流程

### 1. 用户注册
1. 用户访问注册页面 `/register`
2. 输入邮箱并点击「获取验证码」，系统向邮箱发送6位数字验证码（有效期10分钟）
3. 输入验证码、密码和昵称后提交注册
4. 验证码校验通过即创建账号，并直接标记邮箱已验证

### 2. 邮箱验证
针对历史账号或管理员手动触发的情况：
1. 用户点击邮件中的验证链接
2. 跳转到验证页面 `/verify-email?token=xxx`
3. 系统自动验证邮箱
4. 验证成功后跳转到首页

### 3. 重新发送验证邮件
1. 登录用户可以在个人中心重新发送验证邮件
2. 未验证用户会看到提醒组件

## 测试方法

### 1. 本地测试
```bash
# 启动后端
cd backend
go run cmd/server/main.go

# 启动前端
cd frontend
npm run dev
```

### 2. 测试邮箱配置
使用测试邮箱服务：
- Mailtrap (https://mailtrap.io)
- MailHog (本地测试)
- Gmail SMTP (需要应用密码)

### 3. 测试步骤
1. 访问 http://localhost:5174/register
2. 使用测试邮箱注册
3. 查收邮箱验证码，填写后完成注册
4. 如需测试补发流程，可在登录后触发验证邮件并按照链接完成验证

## 安全特性

### 1. 令牌安全
- 注册验证码为6位数字，使用加密随机数生成
- 验证码有效期10分钟
- 邮件验证令牌有效期24小时
- 令牌一次性使用
- 令牌与用户绑定

### 2. 防滥用
- 同一邮箱重复请求验证码会覆盖旧记录
- 限制验证邮件发送频率
- 清理过期令牌
- 防止重复验证

## 错误处理

### 常见错误及解决方案
1. **邮件发送失败**
   - 检查SMTP配置
   - 验证邮箱凭据
   - 检查网络连接

2. **验证链接无效**
   - 确认链接是否过期
   - 检查令牌是否正确
   - 重新发送验证邮件

3. **邮箱已验证**
   - 用户已验证，无需重复操作
   - 提示用户邮箱已验证

## 部署注意事项

### 1. 生产环境配置
- 使用正式的SMTP服务
- 配置正确的域名
- 设置HTTPS

### 2. 邮件模板
- 使用品牌化的邮件模板
- 包含退订链接
- 符合反垃圾邮件法规

### 3. 监控
- 监控邮件发送成功率
- 记录验证失败日志
- 设置告警机制

## 扩展功能

### 1. 邮件模板定制
支持自定义邮件模板，包括：
- 品牌Logo
- 个性化内容
- 多语言支持

### 2. 批量验证
支持批量发送验证邮件给多个用户

### 3. 验证统计
提供验证率统计和分析功能
