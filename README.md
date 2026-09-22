# PromptHub

> AI 提示词资产管理工具 · 全栈项目骨架

PromptHub 是一个面向个人用户的 AI 提示词集中管理平台，提供提示词的**分类存储、AI 调优、版本对比、批量导出**等能力。本仓库是项目的**可运行、不含业务功能**的全栈骨架，目标是为后续按 [PRD](./tasks/design/prd.md) 落地业务开发提供统一规范与基础设施。

## 技术栈

| 层 | 技术选型 |
| --- | --- |
| 前端 | TypeScript + Vue3 + Vite + Element Plus + Pinia + Axios |
| 后端 | Go 1.27 + Gin + GORM + PostgreSQL/MySQL |
| 部署 | Docker + Docker Compose |
| 文档 | 全部中文，写入 `tasks/` 与各模块目录 |

## 目录结构

```
prompt-hub/
├── README.md                 本文件 - 项目总览
├── AGENTS.md                 AI / 协作者协作约定
├── docker-compose.yml        一键编排：postgres + backend + frontend
├── docs/                     架构与运维文档（可选）
├── frontend/                 Vue3 前端
├── backend/                  Go 后端
└── tasks/                    项目任务与设计文档
    ├── design/               设计产物（UI、PRD 等）
    │   └── prd.md            产品需求文档
    ├── ui-design-prompt.md   UI 设计提示词（给 AI 设计工具）
    ├── api.md                REST 接口约定
    ├── database.md           数据库设计与表结构约定
    └── deployment.md         Docker 部署约定
```

frontend / backend / 根目录 三者均有独立的 `README.md` 与 `AGENTS.md`。

## 快速开始

### 方式一：本地开发

```bash
# 1) 前端
cd frontend
npm install
npm run dev              # http://localhost:5173

# 2) 后端（单独终端）
cd backend
go mod download
go run ./cmd/server      # http://localhost:8080
```

### 方式二：一键 Docker 启动

```bash
docker compose up -d
# 前端 → http://localhost:5173
# 后端 → http://localhost:8080
# PG   → localhost:5432
```

## 验证清单（与本骨架版本对应）

| 项 | 命令 | 预期 |
| --- | --- | --- |
| 前端类型检查 | `cd frontend && npm run type-check` | 无错误 |
| 前端生产构建 | `cd frontend && npm run build` | 输出到 `frontend/dist` |
| 后端编译 | `cd backend && go build ./...` | 无错误 |
| 后端 vet | `cd backend && go vet ./...` | 无错误 |
| 后端启动 | `cd backend && go run ./cmd/server` | 监听 8080 |
| 健康检查 | `curl http://localhost:8080/healthz` | 返回 `code: 200` |
| 版本接口 | `curl http://localhost:8080/api/v1/version` | 返回 app/version/env |

## 当前未实现（按 PRD 顺序补齐）

业务功能由后续迭代按 [tasks/design/prd.md](./tasks/design/prd.md) 推进：

- 账户体系（注册/登录/改密）
- 提示词 CRUD / 搜索 / 筛选 / 导出
- 分类管理
- AI 大模型配置
- AI 自动生成 / AI 调优 / 调优历史
- 审计日志

骨架已就绪的"基础设施"：

- ✅ 前端：Vue3 + Vite + TS 模板；Element Plus + Pinia + Axios 集成；路由 / 状态管理 / HTTP 客户端 / 默认布局 / 404 页
- ✅ 后端：Viper 配置加载、Zap 日志、GORM 数据库（不建业务表）、统一响应 / 错误码、Recovery / AccessLog / CORS 中间件、Health 检查
- ✅ 部署：`docker-compose.yml`、前后端独立 `Dockerfile`、Nginx 反向代理到后端 `/api`

## 协作约定

详见 [`AGENTS.md`](./AGENTS.md)（含 commit / 分支 / 评审 / 安全 准则）。
