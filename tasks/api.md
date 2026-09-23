# PromptHub REST 接口约定

> 适用范围：前端 ↔ 后端通信的所有 REST 接口定义、版本演进与错误处理约定。
> 状态：本骨架仅实现 `/healthz` 与 `/api/v1/version`，业务接口按 [PRD §八](./design/prd.md) 增量添加。

---

## 1. 基本约定

- **路径前缀**：`/api/v1`
- **认证**：`Authorization: Bearer <access_token>`（业务接口启用后强制；当前骨架未启用）
- **字符集**：UTF-8
- **请求 / 响应**：`application/json; charset=utf-8`
- **跨域**：开发环境通过 Vite 代理；生产由 Nginx 反向代理到 `backend:8080/api/`

## 2. 统一响应结构

| 字段 | 类型 | 含义 |
| --- | --- | --- |
| `code` | int | 业务码；200 表示成功，详见 §3 |
| `message` | string | 给用户看的中文消息 |
| `data` | any \| null | 业务对象，成功时返回 |

成功示例（健康检查）：

```json
{
  "code": 200,
  "message": "OK",
  "data": {
    "app": "PromptHub",
    "version": "0.1.0",
    "env": "development",
    "timestamp": "2026-09-22T06:13:39Z",
    "deps": { "database": "ok" }
  }
}
```

分页响应：

```json
{
  "code": 200,
  "message": "OK",
  "data": {
    "list": [...],
    "total": 248,
    "page": 1,
    "page_size": 20
  }
}
```

## 3. 业务错误码

定义在 `backend/pkg/errcode/errcode.go`，与 HTTP 状态码解耦：

| 业务码 | 含义 | 默认 HTTP 状态 |
| --- | --- | --- |
| 200 | OK | 200 |
| 400 | 参数错误 | 400 |
| 401 | 未登录 | 401 |
| 403 | 无权限 | 403 |
| 404 | 资源不存在 | 404 |
| 409 | 资源冲突（重复） | 409 |
| 429 | 限流 | 429 |
| 500 | 服务器内部错误 | 500 |
| 502 | 上游错误（AI 模型） | 502 |
| 503 | 服务暂不可用 | 503 |
| 504 | 上游超时 | 504 |

## 4. 当前已实现

| 方法 | 路径 | 描述 | 鉴权 |
| --- | --- | --- | --- |
| GET | `/healthz` | 健康检查（含 DB 连通性） | 否 |
| GET | `/api/v1/version` | 应用版本信息 | 否 |
| POST | `/api/v1/auth/register` | 用户注册，成功后自动签发双 token | 否 |
| POST | `/api/v1/auth/login` | 邮箱+密码登录；统一提示"邮箱或密码错误"（防枚举） | 否 |
| POST | `/api/v1/auth/refresh` | 用 refresh_token 换取新 access_token | 否 |
| POST | `/api/v1/auth/logout` | 登出（本期服务端无副作用，前端清 token） | 是 |
| GET | `/api/v1/prompts` | 提示词列表（分页+搜索+筛选+排序） | 是 |
| POST | `/api/v1/prompts` | 创建提示词 | 是 |
| GET | `/api/v1/prompts/:id` | 提示词详情 | 是 |
| PUT | `/api/v1/prompts/:id` | 更新提示词（增量） | 是 |
| DELETE | `/api/v1/prompts/:id` | 软删提示词（30 天可恢复） | 是 |
| POST | `/api/v1/prompts/batch-delete` | 批量软删提示词 | 是 |
| POST | `/api/v1/prompts/:id/favorite` | 切换收藏标记 | 是 |
| GET | `/api/v1/categories` | 列出当前用户的所有分类 | 是 |
| POST | `/api/v1/categories` | 创建分类 | 是 |
| PUT | `/api/v1/categories/:id` | 更新分类（增量） | 是 |
| DELETE | `/api/v1/categories/:id` | 软删分类；分类下存在提示词时返回 409 | 是 |

## 5. 后续按 PRD 实现的接口（待办）

业务接口按 [PRD §八](./design/prd.md) 章节落地：

| 分组 | 端点 |
| --- | --- |
| 鉴权 | `/auth/register`、`/auth/login`、`/auth/refresh`、`/auth/logout`、`/auth/change-password` |
| 提示词 | `/prompts`、`/prompts/:id`、`/prompts/batch-delete`、`/prompts/export`、`/prompts/:id/favorite`、`/prompts/:id/tune-histories` |
| 分类 | `/categories`、`/categories/:id`、`/categories/reorder` |
| AI 能力 | `/llm/generate` (SSE)、`/llm/tune` (SSE)、`/llm/test-connection` |
| 大模型配置 | `/llm-configs`、`/llm-configs/:id`、`/llm-configs/:id/default` |
| 用户 | `/me`、`/me` (PATCH) |

具体字段、请求体、响应体按 PRD 第五章对应页面的"字段"与"行为"落实。

## 6. 鉴权流程（启用后）

1. 登录返回 `access_token`（短期，Header 携带）+ `refresh_token`（长期，HttpOnly Cookie）。
2. 业务接口从 Header `Authorization` 读取 `access_token`；过期返回 401。
3. 前端静默调用 `/auth/refresh` 续期；续期失败则强制回 `/login`。
4. `change-password` 时 bump `users.token_version`，使旧的 refresh_token 全部失效。

## 7. CORS / 安全头

- 由 `middleware.CORS` 默认挂载；生产环境务必收紧 `Access-Control-Allow-Origin` 到白名单。
- 全站 HTTPS（部署侧由 Nginx / 负载均衡终结 TLS）。
- API Key 类敏感字段（如 `llm_configs.api_key_enc`）绝不返回明文到前端；列表中脱敏（如 `sk-***abc123`）。

## 8. 接口演进

- 重大变更（路径前缀、字段重命名）走版本号升级：`/api/v2`。
- 兼容性变更（新增可选字段）：直接修改 + 在 PR 描述中标注。
