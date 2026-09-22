# PromptHub 数据库设计约定

> 适用范围：所有 GORM 模型、迁移、查询、字段命名、索引约定。
> 状态：当前骨架**未创建任何业务表**。业务表按 [PRD §七](./design/prd.md) 增量加入，并在本文件登记。

---

## 1. 数据库选型

| 驱动 | 适用场景 | 配置键 |
| --- | --- | --- |
| PostgreSQL 14+ | 生产 / 默认 | `db.driver=postgres` |
| MySQL 8.0+ | 可选（见 `docker-compose.yml` 注释段） | `db.driver=mysql` |

> 业务实现需兼容两种驱动；不在 model 层写方言特定 SQL。

## 2. 命名约定

- 表名：小写 + 下划线，前缀 `ph_`（PromptHub）。
- 字段：小写 + 下划线。
- 主键：优先使用 `BIGINT AUTO_INCREMENT`（PostgreSQL：`BIGSERIAL`）；是否迁移到 UUID 由后续决定。
- 必备字段：`created_at`、`updated_at`、`deleted_at`（GORM v2 软删约定）。
- 外键命名：`fk_<table>_<ref_table>`（如 `fk_ph_prompts_ph_categories`），由迁移工具生成。

## 3. 通用字段

所有业务表包含：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | BIGINT / PK | 主键 |
| `created_at` | DATETIME | 创建时间（UTC） |
| `updated_at` | DATETIME | 更新时间（UTC） |
| `deleted_at` | DATETIME NULL | 软删除时间；NULL = 有效 |

> 软删除保留 30 天可恢复；之后由异步任务清理（PRD §九）。

## 4. 多租户 / 数据隔离

- 所有业务表必须携带 `user_id`（与 `ph_users.id` 关联）。
- Repository 强制 `WHERE user_id = ?`；不允许 join / select 跨用户数据。
- 删除分类等敏感操作需先校验 `user_id`。

## 5. 索引约定

| 场景 | 索引 |
| --- | --- |
| 列表分页按时间倒序 | `(user_id, deleted_at, updated_at DESC)` |
| 按分类筛 | `(user_id, category_id)` |
| 按收藏筛 | `(user_id, is_favorite)` |
| 调优历史按版本号 | `(prompt_id, version_no DESC)` |
| 全文搜索（可选） | `FULLTEXT(title, description, content)`（MySQL）/ `GIN`（PostgreSQL） |
| 软删与唯一约束 | 唯一索引在 `WHERE deleted_at IS NULL` 下生效（部分索引） |

## 6. 计划中的业务表（按 PRD §七）

> 本骨架暂不创建；后续迭代先在本文件登记，再创建对应 GORM model。

### 6.1 `ph_users` 用户表

| 字段 | 类型 | 约束 |
| --- | --- | --- |
| id | BIGINT | PK |
| email | VARCHAR(128) | UNIQUE NOT NULL |
| password_hash | VARCHAR(255) | NOT NULL |
| nickname | VARCHAR(20) |  |
| avatar_url | VARCHAR(255) |  |
| token_version | INT | DEFAULT 0 |
| created_at / updated_at / deleted_at | DATETIME |  |

### 6.2 `ph_categories` 分类表

| 字段 | 类型 | 约束 |
| --- | --- | --- |
| id | BIGINT | PK |
| user_id | BIGINT | FK→users NOT NULL |
| name | VARCHAR(20) | NOT NULL（同一用户下唯一） |
| color | VARCHAR(7) | HEX |
| sort | INT | DEFAULT 0 |

唯一索引：`(user_id, name)` WHERE `deleted_at IS NULL`。

### 6.3 `ph_prompts` 提示词表

| 字段 | 类型 | 约束 |
| --- | --- | --- |
| id | BIGINT | PK |
| user_id | BIGINT | FK NOT NULL |
| category_id | BIGINT | FK→categories NULL |
| title | VARCHAR(100) | NOT NULL |
| description | VARCHAR(200) |  |
| content | TEXT | NOT NULL |
| is_favorite | BOOLEAN | DEFAULT false |
| tags | JSON/ARRAY |  |
| version_no | INT | DEFAULT 1 |

### 6.4 `ph_tune_histories` 调优历史表

| 字段 | 类型 | 约束 |
| --- | --- | --- |
| id | BIGINT | PK |
| user_id | BIGINT | FK NOT NULL |
| prompt_id | BIGINT | FK→prompts NOT NULL |
| version_no | INT | NOT NULL |
| request_text | TEXT |  |
| original_content | MEDIUMTEXT | NOT NULL |
| tuned_content | MEDIUMTEXT | NOT NULL |
| model_config_id | BIGINT | FK→llm_configs |
| is_applied | BOOLEAN | DEFAULT false |

### 6.5 `ph_llm_configs` 大模型配置表

| 字段 | 类型 | 约束 |
| --- | --- | --- |
| id | BIGINT | PK |
| user_id | BIGINT | FK NOT NULL |
| name | VARCHAR(20) | NOT NULL |
| provider | VARCHAR(32) | NOT NULL |
| model | VARCHAR(100) | NOT NULL |
| base_url | VARCHAR(255) | NOT NULL |
| api_key_enc | TEXT | NOT NULL（加密） |
| is_enabled | BOOLEAN | DEFAULT true |
| is_default | BOOLEAN | DEFAULT false |
| last_test_status | VARCHAR(16) |  |
| last_test_at | DATETIME |  |

唯一索引：`(user_id, name)` WHERE `deleted_at IS NULL`。

### 6.6 `ph_audit_logs` 操作日志表（V1 可选）

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | BIGINT | PK |
| user_id | BIGINT | FK |
| action | VARCHAR(32) | LOGIN / DELETE_PROMPT / EXPORT / CHANGE_PASSWORD |
| target_type | VARCHAR(32) | prompt / llm_config / user |
| target_id | VARCHAR(64) |  |
| ip | VARCHAR(45) |  |
| ua | VARCHAR(255) |  |

## 7. 安全

- 密码字段使用 `argon2id` 或 `bcrypt`，禁止明文。
- API Key 使用 AES-GCM 对称加密；密钥从环境变量注入。
- 软删除策略：保留 30 天可恢复，超期清理。

## 8. 迁移与种子数据

- 迁移策略：业务模型通过 `Makefile` 中的 `make migrate` 触发；禁止直接在生产数据库手工改结构。
- 种子数据：仅开发环境按需加载（`scripts/seed/`），不进生产镜像。

## 9. 备份与清理

- 数据库容器声明 `postgres_data` 卷（见 `docker-compose.yml`）。
- 备份脚本（`scripts/backup.sh`）默认每周一次全量 + binlog/WAL 增量。
- 删除清理任务：在 `internal/scheduler` 中实现（V1.1+），每日扫描 `deleted_at < NOW() - INTERVAL '30 days'` 物理删除。
