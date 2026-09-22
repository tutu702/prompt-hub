# AGENTS.md - Frontend

> 适用于 `frontend/` 目录下所有 Vue / TS / 样式 / 资源改动。

## 工作原则

1. 仅修改任务要求的内容，不要顺手清理或重构无关代码。
2. 新建文件前确认目录约定（视图 → `views/`、通用组件 → `components/`、业务组件可放在对应业务域子目录）。
3. 中文优先：所有用户可见文案、注释使用中文；技术术语保留英文。
4. 改动完必须可被 `npm run type-check` 与 `npm run build` 通过。

## 编码约定

### Vue 组件

- 统一 `<script setup lang="ts">` 单文件组件。
- 不在模板里写复杂表达式；超过 1 行的逻辑放进 `computed` 或方法。
- Props 用 `defineProps<{ ... }>()` 类型化；Emits 同理。
- 命名：文件 `PascalCase.vue`；导出名同文件名。

### 路由

- 路径前缀：`/api` 后端 API；前端路由按 PRD 章节 4 落地。
- 路由懒加载：`() => import('@/views/...')`。
- 路由 meta：`title` 用于浏览器标题，`requiresAuth` 等业务标记按需添加。

### 状态管理

- 单一职责，按业务域拆 store：`src/stores/xxx.ts`。
- 使用 Composition API 写法（`defineStore('name', () => { ... })`）。
- 状态类型化：`ref<T>('initial')`；避免 `any`。

### HTTP / API

- 所有请求走 `@/api/http`；不在业务层直接 `import axios`。
- 响应解包：`{ code, data, message }`，由拦截器统一处理。
- 业务接口放在 `src/api/modules/`，按业务模块分类（如 `prompt.ts`、`category.ts`）。
- 错误处理：依赖拦截器统一 Toast，业务层只关心结果。

### 样式

- 全局样式集中在 `src/styles/`；组件作用域样式使用 `<style scoped>`。
- 主色 / 间距 / 圆角等设计 token 见 `src/styles/index.scss`；新增 token 同步更新。
- 不在组件里写魔法值（如颜色 / 字号），优先用 CSS 变量或 Element Plus 主题变量。

### 路径别名

- `@/` → `src/`
- 引用顺序：第三方 → 项目内（绝对路径）→ 同目录相对路径

## 变更清单（动到这些文件必须配套更新）

| 变更 | 必须同步 |
| --- | --- |
| 新增 / 修改路由 | 根 README + tasks/api.md |
| 修改 API base 或代理 | `.env.example` + 根 README |
| 新增依赖 | `package.json` + 本 README 命令清单 |

## 验证

```bash
npm run type-check     # TS 严格检查
npm run build          # 生产构建（包含 vue-tsc）
```

任一项失败禁止合并。
