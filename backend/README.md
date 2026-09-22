# PromptHub Backend

> Go 1.27 + Gin + GORM + PostgreSQL/MySQL 的后端骨架。

## 启动

```bash
cp .env.example .env                       # 可选
go mod download
go run ./cmd/server                        # http://localhost:8080
```

数据库默认指向 `127.0.0.1:5432`；本地未连接时服务仍可启动，`/healthz` 中 `database` 会显示 `not configured`。

## 命令

| 命令 | 说明 |
| --- | --- |
| `go run ./cmd/server` | 本地运行 |
| `go build ./...` | 编译所有包（输出二进制到本地） |
| `go vet ./...` | 静态检查 |
| `go mod tidy` | 整理依赖 |
| `go test ./...` | 运行测试（V1 暂无业务测试） |

## 目录

```
backend/
├── cmd/server/             入口 main.go
├── configs/                默认配置文件 config.yaml
├── internal/
│   ├── config/             Viper 配置加载
│   ├── handler/            HTTP 处理层（health 等）
│   ├── middleware/         Recovery / AccessLog / CORS
│   ├── model/              GORM 模型（本骨架不含业务模型）
│   ├── repository/         数据访问层（业务层按需新增）
│   ├── router/             路由注册
│   ├── server/             HTTP 服务组装
│   ├── service/            业务逻辑层
│   └── pkg/
│       ├── database/       GORM 初始化
│       ├── logger/         Zap 初始化
│       └── response/       统一响应
├── pkg/                    对外可用工具
│   ├── errcode/            业务错误码
│   └── utils/              通用工具
├── scripts/                运维脚本
├── go.mod / go.sum
├── Dockerfile
└── .env.example
```

## 配置

配置优先级：**环境变量 (PROMPTHUB_*) > .env > config.yaml > 代码默认值**。

完整字段见 `configs/config.yaml`。常用环境变量：

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| `PROMPTHUB_SERVER_PORT` | `8080` | 监听端口 |
| `PROMPTHUB_SERVER_MODE` | `debug` | gin 模式：debug / release / test |
| `PROMPTHUB_LOG_LEVEL` | `info` | debug / info / warn / error |
| `PROMPTHUB_LOG_FORMAT` | `console` | console / json |
| `PROMPTHUB_DB_DRIVER` | `postgres` | postgres / mysql |
| `PROMPTHUB_DB_HOST` | `127.0.0.1` | 数据库主机 |
| `PROMPTHUB_DB_PORT` | `5432` | 数据库端口 |
| `PROMPTHUB_DB_USER` | `postgres` | 用户名 |
| `PROMPTHUB_DB_PASSWORD` | `postgres` | 密码 |
| `PROMPTHUB_DB_DBNAME` | `prompthub` | 数据库名 |
| `PROMPTHUB_JWT_SECRET` | `change-me-in-production` | JWT 密钥（生产务必替换） |

## 接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/healthz` | 健康检查（含数据库连通） |
| GET | `/api/v1/version` | 应用版本 |
| GET | `/api/v1/...` | 业务接口（按 PRD 后续添加） |

健康检查响应示例：

```json
{
  "code": 200,
  "message": "OK",
  "data": {
    "app": "PromptHub",
    "version": "0.1.0",
    "env": "development",
    "timestamp": "2026-09-22T06:13:39Z",
    "deps": { "database": "ok | fail: <原因> | not configured" }
  }
}
```

详见 [`../AGENTS.md`](../AGENTS.md) 与本目录 [`AGENTS.md`](./AGENTS.md)。
