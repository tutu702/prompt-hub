/**
 * 应用入口文件
 * 负责创建 Vue 应用实例、注册 Pinia / Vue Router / Element Plus 等核心能力。
 */
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import 'element-plus/dist/index.css'

import App from './App.vue'
import router from './router'
import { setUserStoreGetter } from './api/http'
import { useUserStore } from './stores/user'
import './styles/index.scss'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(ElementPlus, { locale: zhCn })

// 把 Pinia store getter 注入到 http 客户端，让请求拦截器能拿到 access token 注入 Authorization
setUserStoreGetter(() => useUserStore())

app.mount('#app')
