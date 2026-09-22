# AGENTS.md - Backend

> 适用于 `backend/` 目录下所有 Go 代码、配置、依赖、Docker 镜像改动。

## 工作原则

1. 包分层：`handler → service → repository → model`，上层依赖下层，禁止反向。
2. 不在 handler 层写业务逻辑；不在 service 层直接读写 `*gin.Context`。
3. 配置通过 `internal/config` 加载；不要在 handler/service 里 `os.Getenv`。
4. 错误：业务错误用 `pkg/errcode` 的常量；HTTP 层使用 `response.Fail` 系列。
5. 改动完必须可被 `go build ./...` 与 `go vet ./...` 通过。

## 编码约定

### 命名

- 包名：小写、单词；避免下划线 / 中划线。
- 文件名：snake_case。
- 结构体：PascalCase；接口以行为命名（`Reader` / `Authenticator`）。
- 错误变量：`ErrXxx`；构造函数：`NewXxx`。

### 日志

- 入口函数注入 `*zap.Logger`，不要用全局 logger。
- 关键路径（启动、关闭、错误、外部调用）必须打日志。
- 不要在日志里输出敏感信息（密码、token、API Key 全文）。

### 数据库

- 业务模型放 `internal/model`，按"实体名 + 字段"风格命名。
- 数据访问放 `internal/repository`：每个实体一个 repo，暴露 `Create / Update / Delete / Get / List`。
- Service 通过接口注入 Repository，便于单元测试 mock。
- 事务：在 Repository 层提供 `WithTx(ctx, fn func(tx *gorm.DB) error) error`，业务层不直接持有 `*gorm.DB`。

### HTTP 层

- handler 函数签名：`func (h *XxxHandler) Method(c *gin.Context)`。
- 输入参数：`c.ShouldBindJSON / ShouldBindQuery / ShouldBindUri`，校验失败用 `response.Fail(c, http.StatusBadRequest, errcode.CodeBadRequest, msg)`。
- 路径前缀：业务接口全部 `/api/v1/...`；Health 在根 `/healthz`。
- 中间件：跨域、Recovery、AccessLog 已默认挂载。新增中间件在 `internal/middleware` 集中。

### 错误处理

- 业务异常转为 `errcode.CodeXxx` + 自定义消息。
- 未捕获的 panic 由 `middleware.Recovery` 兜底返回 500。

## 变更清单

| 变更 | 必须同步 |
| --- | --- |
| 新增 / 修改路由 | `internal/router/router.go` + 根目录 `tasks/api.md` |
| 新增 / 修改配置字段 | `internal/config/config.go` + `configs/config.yaml` + `.env.example` + 本 README |
| 新增依赖 | `go.mod` / `go.sum`（必须跑 `go mod tidy`） |
| 修改镜像构建 | 根目录 `docker-compose.yml` + `tasks/deployment.md` |

## 验证

```bash
go build ./...
go vet ./...
go run ./cmd/server &
curl http://localhost:8080/healthz
```

任一项失败禁止合并。
