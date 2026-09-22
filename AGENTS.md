# AGENTS.md - PromptHub 协作约定

> 适用于：参与本项目的所有 AI 协作者（Claude Code / Codex / 其他 Agent）与人工开发者。
> 文档目的：明确工程规范、协作礼仪、变更流程，避免无序修改。

---

## 一、基本原则

1. **文档先行**：所有跨模块改动，必须先对齐 `tasks/` 内的设计文档，再动代码。
2. **最小改动**：仅做必要的修改；不要顺手重构未涉及的代码。
3. **中文优先**：所有用户可见文案、注释、提交说明默认中文（必要处保留英文术语）。
4. **不破坏骨架**：当前阶段的项目骨架（frontend / backend）仅保证"可运行 + 基础设施完整"，业务功能按 PRD 增量添加，禁止随意删除或重命名核心目录。
5. **可验证**：每次提交应配套可复现的验证步骤（命令 + 预期结果）。

---

## 二、目录与权责

| 路径 | 责任 |
| --- | --- |
| `frontend/` | 前端代码，任何 Vue / TS / 样式 / 路由改动落在该目录 |
| `backend/` | 后端代码，任何 Go / 配置 / 中间件 / 路由改动落在该目录 |
| `tasks/design/` | PRD、设计稿、设计说明等设计产物 |
| `tasks/api.md` | REST 接口约定；改后端路由或新增端点时同步更新 |
| `tasks/database.md` | 数据库 / GORM 模型约定；新增表 / 修改字段时同步更新 |
| `tasks/deployment.md` | Docker 部署约定；改 `docker-compose.yml` / `Dockerfile` 时同步更新 |
| `docs/` | 架构图、运维手册等辅助说明（可选） |

> 每个子目录（`frontend/`、`backend/`、根目录）**均有自己的 README.md 与 AGENTS.md**；子模块的细则以子模块文档为准。

---

## 三、命名与编码约定

### 通用

- 文件名：源码 `kebab-case` 或 `snake_case`（与生态保持一致）；Vue 组件 `PascalCase.vue`；Go 文件 `snake_case.go`。
- 缩进：前端 2 空格；后端 `gofmt` 制表符。
- Linting：前端 `vue-tsc`、后端 `go vet`；CI 阶段加入 ESLint / golangci-lint（如需请先在本文件登记）。
- 提交：`feat:` / `fix:` / `refactor:` / `docs:` / `chore:` / `test:` 前缀。

### 前端

- 路由懒加载：`component: () => import('@/views/...')`。
- 状态：Pinia 单一职责，按业务域拆 `stores/xxx.ts`。
- HTTP：所有请求走 `src/api/http.ts`，统一处理响应解包与错误。
- 组件：通用组件放 `src/components/`，业务组件放业务目录的 `components/` 子目录。
- 图标：默认按需引入 Element Plus Icons，命名 `<el-icon><Edit /></el-icon>`。

### 后端

- 包分层：`handler → service → repository → model`；上层调用下层，禁止反向依赖。
- 错误：`pkg/errcode` 中的业务码；handler 层只负责协议转换，不写业务逻辑。
- 配置：通过 `internal/config` 加载，**禁止在 handler / service 中直接读环境变量**。
- 日志：使用 `internal/pkg/logger` 的 zap 实例；不要引入额外日志库。
- 数据库：`internal/pkg/database` 提供 `*gorm.DB`；业务模型迁移在 `internal/model` 显式注册，本骨架不执行 AutoMigrate。

---

## 四、变更流程

1. **小改动**（单文件、单模块）：直接修改并在 commit 中说明。
2. **跨模块改动**（如新增业务接口涉及前后端 + 数据库 + 部署）：
   1. 先更新 `tasks/design/prd.md` 与 `tasks/api.md` / `tasks/database.md` / `tasks/deployment.md`。
   2. PR 描述里贴出对应的设计章节与 diff。
   3. 同步通知前端 / 后端 owner 评审。
3. **破坏性改动**（删字段、改表名、改路径前缀）：必须升级主版本号或走 CHANGELOG 记录。

---

## 五、安全与隐私

- 禁止在前端 / 日志中明文打印：API Key、密码、token、明文个人身份信息。
- API Key 仅在后端解密后使用，传输到前端前必须脱敏。
- 数据库密码、密钥等敏感配置走环境变量，不入 git。
- SQL 全部走 GORM，不拼接字符串。

---

## 六、与 AI 协作时的特别约定

- 仅修改任务明确要求的内容；不擅自"顺手"扩展。
- 尽量使用项目已有模式（目录、文件结构、命名）；新增文件先确认有无近似模板。
- 每次较大改动后给出 1 段中文"变更摘要"。
- 不确定时使用 `AskUserQuestion` 而非猜测。
