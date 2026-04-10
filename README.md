# 杭州电子科技大学食物点评系统

该项目实现了一个面向杭电师生的食物点评平台，包含 **Go 后端** 与 **React + Vite + Ant Design 前端**。用户可注册登录、提交点评、上传图片，管理员审核后方可对外展示，并支持分页、关键字搜索与多维排序。

## 目录结构

```
.
├── backend/            # Go 后端服务
│   ├── cmd/server      # 程序入口
│   └── internal/       # 配置、数据库、业务逻辑、路由、存储等
├── frontend/           # React 前端应用（Vite）
├── docs/architecture.md# 架构设计说明
├── Makefile            # 常用命令
└── README.md
```

## 本地运行

### 后端
1. 进入 `backend` 目录：`cd backend`
2. 设置必要的环境变量（至少需要 JWT 密钥）：
   ```bash
   export APP_AUTH_JWT_SECRET="change-me"
   # 如需初始化管理员，可同时设置：
   export APP_ADMIN_EMAIL="admin@example.com"
   export APP_ADMIN_PASSWORD="Admin123"
   ```
3. 启动服务：
   ```bash
   go run ./cmd/server
   ```
4. 服务默认监听 `http://localhost:8080`，数据默认使用 `data/app.db`（SQLite）。

> 提示：也可以在仓库根目录直接运行 `make backend`（会注入开发用密钥）。

### 前端
1. 进入 `frontend` 目录：`cd frontend`
2. 安装依赖：`npm install`
3. 运行开发服务：`npm run dev`
4. 打开浏览器访问 `http://localhost:5173`

开发服务器通过 Vite 代理把 `/api` 请求转发到后端，确保两个服务同时启动即可完成联调。

### 小程序
1. 进入 `mini-program/hdu-comment-mini` 目录：`cd mini-program/hdu-comment-mini`
2. 安装依赖：`npm install`
3. 本地开发模式运行：
   ```bash
   npm run dev:weapp
   ```
4. 如果需要让真机或开发者工具访问局域网后端，启动时覆盖接口域名：
   ```bash
   TARO_APP_API_BASE_URL=http://192.168.1.20:8080/api/v1 \
   TARO_APP_ASSET_BASE_URL=http://192.168.1.20:8080 \
   npm run dev:weapp
   ```
5. 构建测试包或正式包：
   ```bash
   TARO_APP_API_BASE_URL=https://staging-api.example.com/api/v1 \
   TARO_APP_ASSET_BASE_URL=https://staging-api.example.com \
   TARO_APP_ENABLE_DEBUG=true \
   npm run build:weapp:test

   TARO_APP_API_BASE_URL=https://api.example.com/api/v1 \
   TARO_APP_ASSET_BASE_URL=https://api.example.com \
   TARO_APP_ENABLE_DEBUG=false \
   npm run build:weapp:prod
   ```

小程序配置采用 `config/index.ts + config/dev.ts + config/prod.ts` 三层结构：
- `config/index.ts`：公共 Taro 配置
- `config/dev.ts`：开发环境默认值，适合本地或局域网联调
- `config/prod.ts`：生产构建默认值，可通过环境变量覆盖

小程序常用环境变量：
- `TARO_APP_API_BASE_URL`：后端 API 地址，例如 `https://api.example.com/api/v1`
- `TARO_APP_ASSET_BASE_URL`：图片和静态资源访问前缀，例如 `https://cdn.example.com`
- `TARO_APP_ENV`：环境标识，例如 `development` / `staging` / `production`
- `TARO_APP_ENABLE_DEBUG`：是否开启调试开关，`true` 或 `false`

> 注意：正式发布到微信小程序时，请使用 HTTPS 域名，并在微信公众平台配置 request / upload / download 合法域名。

## Docker 部署

项目提供了基于 Docker 的一键启动方案：

1. 确保已安装 Docker 与 Docker Compose。
2. 在仓库根目录复制示例配置：
   ```bash
   cp docker-compose.example.yml docker-compose.yml
   ```
3. 按你的环境修改本地 `docker-compose.yml` 或新增 `docker-compose.override.yml` / `.env`。
4. 启动服务：
   ```bash
   docker compose build
   docker compose up -d
   ```
5. 服务启动后：
   - 后端 API 暴露在 `http://localhost:8080`
   - 前端页面可通过 `http://localhost:5173` 访问

默认使用 SQLite 存储，数据与上传的图片会持久化到 Compose 定义的卷 `backend-data` 与 `backend-uploads` 中。仓库中的 [docker-compose.example.yml](/Users/yinchangyu/code/hdu-dp/hdu-comment/docker-compose.example.yml) 仅作为模板，请不要直接在仓库版本里写入生产密钥或服务器私有配置。

停止并清理容器：

```bash
docker compose down
```

## 核心功能
- **用户管理**：邮箱注册登录、QQ OAuth 登录、手机短信验证码登录、个人信息查询。注册成功自动获取登录态（JWT）。
- **点评提交**：上传食物名称、地址、描述、评分；支持追加图片，可配置本地文件或 S3/OSS/COS 等对象存储。
- **审核流程**：管理员查看待审核点评，支持通过或驳回并记录原因。普通用户仅能查看已审核内容和自己的历史提交。
- **公开浏览**：无需登录即可浏览已审核点评详情及图片；支持分页、关键字搜索及按评分/时间排序。
- **令牌刷新**：后端提供访问令牌 + 刷新令牌，前端自动处理 401 并刷新会话。

## 重要配置项
- `APP_SERVER_PORT`：服务端口，默认 `8080`
- `APP_SERVER_READ_HEADER_TIMEOUT` / `APP_SERVER_READ_TIMEOUT` / `APP_SERVER_WRITE_TIMEOUT` / `APP_SERVER_IDLE_TIMEOUT`：HTTP 服务超时配置
- `APP_SERVER_SHUTDOWN_TIMEOUT`：优雅退出超时时间，默认 `10s`
- `APP_LOG_LEVEL`：日志级别，支持 `debug` / `info` / `warn` / `error`
- `APP_LOG_FORMAT`：日志格式，支持 `text` / `json`
- `APP_DATABASE_DSN`：数据库 DSN，默认 `file:data/app.db?_fk=1&mode=rwc`
- `APP_AUTH_JWT_SECRET`：JWT 密钥（必填）
- `APP_AUTH_REFRESH_TOKEN_TTL`：刷新令牌有效期，默认 `168h`
- `APP_AUTH_QQ_ENABLED`：是否启用 QQ 登录（默认 `false`）
- `APP_AUTH_QQ_APP_ID` / `APP_AUTH_QQ_APP_SECRET` / `APP_AUTH_QQ_REDIRECT_URI`：QQ OAuth 配置（启用 QQ 登录时必填）
- `APP_AUTH_SMS_ENABLED`：是否启用短信登录（默认 `false`）
- `APP_AUTH_SMS_CODE_TTL`：短信验证码有效期（默认 `10m`）
- `APP_AUTH_SMS_DEV_MODE`：短信开发模式（默认 `true`，验证码会写入后端日志并通过接口 `debug_code` 返回）
- `APP_STORAGE_PROVIDER`：存储类型，`local`（默认）或 `s3`
  - Local 模式：
    - `APP_STORAGE_UPLOAD_DIR`：图片物理存储目录，默认 `uploads`
    - `APP_STORAGE_PUBLIC_BASE_URL`：图片访问前缀，默认 `/api/v1/uploads`
  - S3/OSS/COS 模式（兼容 S3 协议）：
    - `APP_STORAGE_S3_ENDPOINT`
    - `APP_STORAGE_S3_BUCKET`
    - `APP_STORAGE_S3_REGION`
    - `APP_STORAGE_S3_ACCESS_KEY`
    - `APP_STORAGE_S3_SECRET_KEY`
    - `APP_STORAGE_S3_USE_SSL`（默认 `true`）
    - `APP_STORAGE_S3_BASE_URL`（可选，若不配置将基于 endpoint 构造）
- `APP_ADMIN_EMAIL` / `APP_ADMIN_PASSWORD`：设置后，会自动创建管理员账号

**分页与搜索参数（示例）：**

```
GET /api/v1/reviews?page=1&page_size=12&query=鸡排&sort=rating&order=desc
```

- `page` / `page_size`：分页
- `query`：在标题、地址、描述中模糊搜索
- `sort`：`created_at` 或 `rating`
- `order`：`asc` / `desc`

所有列表接口（公开列表、我的点评、管理员待审核）均支持上述参数，并返回：

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 42,
    "total_pages": 5
  }
}
```

## 后续扩展建议
1. **完善测试**：为服务层与 Handler 编写单元测试，提高回归信心。
2. **多媒体支持**：扩展为多图上传、视频或富文本点评。
3. **运营工具**：支持后台批量审核、数据导出或看板统计。
4. **通知体系**：结合邮件/企业微信推送审核结果。

## API 文档

项目已集成 Swagger UI，启动服务后可通过浏览器访问：

```
http://localhost:8080/swagger/index.html
```

### 使用说明
- 所有 API 端点都已文档化，包含请求参数、响应格式和示例
- 支持在线测试 API 功能
- 文档自动生成，与代码保持同步

### 开发指南
如需更新 API 文档，请在 `backend` 目录执行：
```bash
swag init
```

欢迎继续完善并部署到生产环境！
