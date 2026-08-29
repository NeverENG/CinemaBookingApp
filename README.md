# LTerm 电影院订票系统

Go(Gin) + PostgreSQL 后端，React + TypeScript + ECharts 前端。架构对齐 potentia/backend：`core/domain` + `core/port` + `infra` + `app`，详见 [docs/guide/LTerm后端结构蓝图.md](docs/guide/LTerm后端结构蓝图.md)；前端 UIUX 与分层架构见 [docs/design/07-前端UIUX与架构.md](docs/design/07-前端UIUX与架构.md)。

## 快速开始

### 本机（Homebrew PostgreSQL）

```bash
brew services start postgresql@16
createdb cinema        # 首次
make migrate           # 建表 + 种子数据 + 引导账号
make run               # 启动 :8080
```

默认引导账号：`admin/admin123`（管理员）、`demo/demo123`（用户）。可用 `.env` 或环境变量覆盖，变量清单见 `.env.example`。

### Docker Compose

```bash
make compose-up        # 启动 postgres(5433) + backend(8080)，自动迁移
make compose-down
```

## 常用命令

```bash
make test              # 单元测试
make test-integration  # 集成测试（需 TEST_DB 库存在）
make build             # 输出 bin/lterm
make vet
```

## 前端

```bash
cd frontend
npm install
VITE_API_BASE_URL=http://localhost:8080/api/v1 npm run dev
```

前后端联调链路已覆盖：电影/影院查询、场次与实时座位图、登录、锁座下单、Mock 支付回调、订单详情/列表和退款。后端不可用时，用户端仅对网络错误降级为演示数据；业务错误会直接反馈。

## 目录速览

```text
cmd/lterm           入口
internal/core       领域：domain（实体/状态机）+ port（接口）
internal/infra      实现：config / database/postgres / ...
internal/app        编排：server/http + server/service + server/biz + job
sql/migrations      迁移（增量文件，禁止改已执行过的）
```
