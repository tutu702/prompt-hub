# PromptHub 架构总览

> 配套 `tasks/design/prd.md` 的整体架构说明。本骨架对应 §1.2 "产品形态"。
> 当前文档为占位，详细设计在后续迭代按业务模块展开。

## 1. 系统分层

```
┌─────────────────────────────────────────────────────────────┐
│  浏览器（PC 优先 / 响应式）                                  │
│  Vue3 + Vite + Element Plus + Pinia + Axios                 │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTPS / JSON
                         ▼
┌─────────────────────────────────────────────────────────────┐
│  反向代理（Nginx）                                           │
│  - SPA fallback                                              │
│  - /api 反代到 backend:8080                                  │
└────────────────────────┬────────────────────────────────────┘
                         ▼
┌─────────────────────────────────────────────────────────────┐
│  后端服务（Go + Gin）                                        │
│  cmd/server                                                  │
│  ├── internal/config  Viper 加载                             │
│  ├── internal/server  HTTP 装配                              │
│  ├── internal/router  路由注册                               │
│  ├── internal/middleware  Recovery / AccessLog / CORS        │
│  ├── internal/handler 请求处理（业务接口）                   │
│  ├── internal/service  业务编排                              │
│  ├── internal/repository 数据访问                            │
│  ├── internal/model GORM 模型                                │
│  └── internal/pkg database / logger / response               │
└────────────────────────┬────────────────────────────────────┘
                         ▼
┌─────────────────────────────────────────────────────────────┐
│  数据存储                                                    │
│  - PostgreSQL 16（默认）/ MySQL 8.4（可选）                  │
│  - 软删除 30 天可恢复                                        │
│  - 所有业务表带 user_id，数据强隔离                           │
└─────────────────────────────────────────────────────────────┘
```

## 2. 模块边界

- **前端**：所有用户交互、UI 渲染、表单校验、状态管理；不直接访问数据库。
- **后端**：鉴权、业务编排、数据校验、数据库读写、AI 调用代理。
- **数据库**：唯一数据来源。
- **AI**：作为外部依赖，通过后端 `internal/pkg/llm`（待新增）适配多种供应商。

## 3. 核心数据流

1. 用户登录 → 后端校验 → 颁发 access + refresh token。
2. 前端请求业务接口 → 携带 access_token → 后端鉴权中间件 → handler → service → repository → DB。
3. AI 调优：前端发起 SSE → 后端转发给 LLM 提供商 → 流式回写到前端 → 保存为调优历史。

## 4. 可观测性

- 应用日志：zap JSON → stdout → Docker 日志收集。
- 健康检查：`/healthz` 含数据库连通性。
- 错误码：详见 `tasks/api.md §3`。

## 5. 安全

- HTTPS 在 Nginx / 负载均衡层终结。
- JWT 短期 + Refresh；改密强制 token_version +1。
- API Key 加密存储；前端展示脱敏。
- 详细约定见 `backend/AGENTS.md` 与 `frontend/AGENTS.md`。
