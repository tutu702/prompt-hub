# PromptHub Docker 部署约定

> 适用范围：本地一键部署、CI 构建、线上单机 / 多机部署。
> 当前骨架提供 `docker-compose.yml` + 前 / 后端 `Dockerfile`，覆盖本地与单机场景。

---

## 1. 服务拓扑

```
┌──────────────────────────────────────────────────────────┐
│                       Docker Compose                     │
│                                                          │
│  ┌──────────────┐    ┌──────────────┐    ┌────────────┐  │
│  │  prompthub-  │ -> │  prompthub-  │ <- │  prompthub-│  │
│  │  frontend    │    │  backend     │    │  postgres  │  │
│  │ (nginx:80)   │    │ (go:8080)    │    │ (postgres  │  │
│  └──────┬───────┘    └──────┬───────┘    │  16:5432)  │  │
│         │                  │            └────────────┘  │
│     localhost:5173     localhost:8080       localhost:5432│
└──────────────────────────────────────────────────────────┘
```

- 前端容器：`prompthub-frontend` 提供 Nginx 静态站点 + 反代 `/api` 到 `backend:8080`。
- 后端容器：`prompthub-backend` 提供 HTTP API（监听 8080）。
- 数据库容器：`prompthub-postgres`（PG 16-alpine）。

## 2. 文件清单

```
prompt-hub/
├── docker-compose.yml
├── frontend/
│   ├── Dockerfile
│   ├── nginx.conf
│   └── .dockerignore
├── backend/
│   ├── Dockerfile
│   └── .dockerignore
└── .env.example
```

## 3. 一键启动

```bash
cp .env.example .env          # 可选：自定义密码、端口
docker compose up -d
docker compose ps
docker compose logs -f prompthub-backend
```

- 前端：`http://localhost:5173`
- 后端：`http://localhost:8080/healthz`
- 数据库：`localhost:5432`

停服 / 清理：

```bash
docker compose down           # 停止并移除容器
docker compose down -v        # 同时清理数据卷（重置）
```

## 4. 镜像构建

### 4.1 后端多阶段构建

- Builder：`golang:1.27-alpine`，`CGO_ENABLED=0` 静态编译，`-trimpath -ldflags "-s -w"`。
- Runtime：`alpine:3.20`，仅安装 `ca-certificates tzdata`，非 root 运行。
- 构建缓存：分层复制 `go.mod / go.sum` 先 `go mod download`，最大化缓存命中。

### 4.2 前端多阶段构建

- Builder：`node:22-alpine`，`npm ci` → `npm run build`。
- Runtime：`nginx:1.27-alpine`，静态文件由 builder 复制到 `/usr/share/nginx/html`。
- Nginx 配置：`frontend/nginx.conf` 实现 SPA fallback + `/api` 反向代理。
- 健康检查：`wget http://127.0.0.1/healthz`。

## 5. 网络与发现

- 所有服务在同一 `bridge` 网络（`prompthub_default`，由 compose 自动创建）。
- 容器间通过服务名解析：`backend`、`frontend`、`postgres`。
- 数据库主机使用 `postgres` 而非 `127.0.0.1`（让容器 DNS 解析到数据库容器）。

## 6. 端口与暴露

| 服务 | 容器端口 | 宿主机端口 | 备注 |
| --- | --- | --- | --- |
| frontend | 80 | 5173 | 可按需修改 |
| backend | 8080 | 8080 | 可按需修改 |
| postgres | 5432 | 5432 | 生产建议改为不暴露，仅内网访问 |

## 7. 数据持久化

| 卷名 | 挂载点 | 说明 |
| --- | --- | --- |
| `postgres_data` | `/var/lib/postgresql/data` | PG 数据持久化 |
| `mysql_data`（可选） | `/var/lib/mysql` | MySQL 数据（默认注释） |

> 切换到 PostgreSQL 时保留 `postgres_data`；切换到 MySQL 时同步替换并清理 `postgres_data`。

## 8. 环境变量约定

所有敏感 / 环境相关参数通过环境变量注入，命名遵循 `PROMPTHUB_*` 前缀（后端）与 `VITE_*`（前端）。详见：

- 后端：[`backend/README.md`](../backend/README.md)
- 前端：[`frontend/README.md`](../frontend/README.md)
- 全栈：[`../.env.example`](../.env.example)

> ❗ 生产环境务必替换 `PROMPTHUB_JWT_SECRET`、`POSTGRES_PASSWORD`、各 `ph_llm_configs.api_key_enc` 加密密钥。

## 9. 镜像加速（可选）

国内构建可在对应 Dockerfile 顶部取消以下注释：

```dockerfile
# ENV GOPROXY=https://goproxy.cn,direct
# RUN npm config set registry https://registry.npmmirror.com
```

## 10. CI / CD 流程（建议）

1. PR 触发：`frontend type-check + build`、`backend go vet + build`。
2. 合并 main 后构建并推送镜像：`ghcr.io/<org>/prompthub-backend:<sha>`、`...frontend:<sha>`。
3. 部署机拉取最新镜像 + `docker compose pull && docker compose up -d`。
4. 健康检查通过后再切流量（Nginx upstream 灰度）。

## 11. 监控与日志

- 应用日志：开发 console 格式；生产建议 `format=json` 接入 Loki / ELK。
- 容器日志：`docker compose logs`；生产接入集中日志系统。
- 健康探针：`/healthz`（业务 + DB），容器自带 `HEALTHCHECK`。

## 12. 后续演进

- 反向代理 + TLS：上线 Nginx / Traefik 自动签发证书。
- 多机部署：拆 `db` 为独立 stack；前后端用 Kubernetes / Swarm。
- 备份：增加定时 `pg_dump` 到对象存储。
- 高可用：PG 改主从；前端接入 CDN。
