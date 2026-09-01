# LTerm 电影院订票系统

Go（Gin + GORM + PostgreSQL）后端，React + TypeScript 前端。项目包含用户端、管理端、多角色路由、选座下单、Mock 支付、订单、退款、积分和票房看板。

## Docker 一键启动

需要安装 Docker Engine 与 Docker Compose Plugin。

```bash
cp .env.example .env

# 部署前至少修改 POSTGRES_PASSWORD、JWT_SECRET 和演示账号密码
docker compose up -d --build
docker compose ps
```

默认访问地址：

```text
前端：http://localhost:8088
健康检查：http://localhost:8088/healthz
PostgreSQL：仅绑定 127.0.0.1:5433
```

`APP_PORT` 和 `POSTGRES_PORT` 可以在 `.env` 中修改。

### 启动顺序与 Seed

Compose 按下面的顺序启动：

```text
PostgreSQL 健康
→ migrate 执行 /app/lterm -migrate -seed
→ 创建默认管理员和演示用户
→ 注入电影、影院、场次、座位、历史订单和票房 Seed
→ backend 健康
→ frontend Nginx 启动
```

Seed 可重复执行，会重建 `OSEED*` 演示订单和演示票房数据：

```bash
docker compose run --rm migrate -seed
```

Seed 会使用 `DEMO_USERNAME` 对应的演示用户；修改 `.env` 后重新执行上述命令即可切换演示账号。

完整重置数据库：

```bash
docker compose down -v --remove-orphans
docker compose up -d --build
```

### 烟测

启动完成后执行：

```bash
./scripts/smoke.sh http://127.0.0.1:8088
```

服务器上将地址替换为服务器 IP 或域名：

```bash
./scripts/smoke.sh http://SERVER_IP:8088
```

烟测会验证健康检查、首页数据、电影列表和演示用户登录。

### 日志与停止

```bash
docker compose logs -f migrate backend frontend
docker compose down
```

后端已处理 `SIGTERM`，Compose 停止或更新容器时会优雅停止 HTTP 服务和定时任务。

## 服务器部署

```bash
git clone <repository-url> LTerm
cd LTerm
cp .env.example .env

# 修改 .env，开放 APP_PORT 对应的防火墙端口
docker compose up -d --build
docker compose ps
./scripts/smoke.sh http://SERVER_IP:8088
```

前端容器通过 Nginx 将 `/api/*` 代理到后端，因此浏览器只访问一个端口，不需要额外配置跨域。需要 HTTPS 时，可以在服务器已有的 Nginx、Caddy 或云负载均衡前配置域名和证书。

更新版本：

```bash
git pull
docker compose up -d --build
```

数据库保存在 Docker Volume `pgdata` 中，普通更新不会删除数据。

## 本机开发

### 后端

```bash
brew services start postgresql@16
createdb cinema
make migrate
make run
```

默认演示账号可以通过 `.env` 修改：

```text
平台管理员：admin / admin123
影院管理员：cinema_admin / cinema123（默认绑定影院 1）
用户：demo@lterm.test / demo123
```

### 邮箱验证码

用户注册和忘记密码均使用 6 位邮箱验证码。使用 QQ 邮箱时，在 `.env` 配置：

```text
SMTP_HOST=smtp.qq.com
SMTP_PORT=465
SMTP_USER=你的QQ邮箱
SMTP_PASSWORD=QQ邮箱设置中生成的SMTP授权码
SMTP_FROM=你的QQ邮箱
```

未配置 `SMTP_USER` / `SMTP_PASSWORD` 时，接口会返回开发验证码，前端会自动填入，便于本机和课程演示。

### 前端

```bash
cd frontend
npm ci
VITE_API_BASE_URL=http://localhost:8080/api/v1 npm run dev
```

## 测试

```bash
go test ./...
go vet ./...
go build -o bin/lterm ./cmd/lterm

cd frontend
npm run build
```

## 常用 Make 命令

```bash
make migrate          # 本机迁移并注入 Seed
make seed             # 本机只重建 Seed
make test
make vet
make compose-up
make compose-ps
make compose-logs
make smoke
make compose-down
make compose-reset    # 删除容器和数据库 Volume
```

## 目录

```text
cmd/lterm           程序入口
internal/core       领域实体、状态机和 Port
internal/infra      配置与 PostgreSQL 实现
internal/app        HTTP、Service、Biz 和 Job
frontend            React 前端与 Nginx 镜像
sql/migrations      数据库迁移和演示 Seed
scripts/smoke.sh    部署后烟测
```
