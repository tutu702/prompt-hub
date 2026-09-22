# PromptHub Frontend

> Vue3 + TypeScript + Vite + Element Plus + Pinia + Axios 的前端骨架。

## 启动

```bash
npm install
cp .env.example .env.local        # 可选，按需修改
npm run dev                       # http://localhost:5173
```

## 命令

| 命令 | 说明 |
| --- | --- |
| `npm run dev` | 启动开发服务器，监听 `0.0.0.0:5173` |
| `npm run build` | 类型检查 + 生产构建，输出 `dist/` |
| `npm run preview` | 本地预览生产产物 |
| `npm run type-check` | 仅执行 `vue-tsc` 类型检查 |

## 目录

```
src/
├── api/            axios 封装 + 各业务模块接口（src/api/modules/*）
├── assets/         图片 / 字体等静态资源
├── components/     通用组件
├── layouts/        布局组件（DefaultLayout 等）
├── main.ts         应用入口
├── App.vue         根组件
├── router/         Vue Router 路由
├── stores/         Pinia 状态
├── styles/         全局样式
├── types/          TS 类型 + auto-import 类型
├── utils/          通用工具函数
└── views/          页面视图
```

## 约定

- 路径别名：`@/` 指向 `src/`。
- API 请求统一走 `@/api/http`，响应解包格式：`{ code, data, message }`。
- 业务 store 按业务域拆，单文件单 store。
- 新组件使用 `<script setup lang="ts">`。

## 环境变量

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| `VITE_API_BASE_URL` | `/api` | 运行时 API 基址 |
| `VITE_API_PROXY_TARGET` | `http://localhost:8080` | 开发代理目标后端 |
| `VITE_APP_TITLE` | `PromptHub` | 应用标题 |

详见 [`../AGENTS.md`](../AGENTS.md) 与本目录 [`AGENTS.md`](./AGENTS.md)。
