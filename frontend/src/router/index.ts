/**
 * Vue Router 路由配置
 *
 * 路由 meta 约定：
 *   - title:    浏览器标题
 *   - layout:   'blank' | 'default'，决定使用哪个 layout 包裹（由 App.vue 切换）
 *   - requiresAuth: true 时必须登录才能访问
 */
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Root',
    // 真正的跳转在 beforeEach 里根据登录态动态决定
    component: () => import('@/views/Home.vue'),
    meta: { title: '首页', layout: 'default', requiresAuth: false },
  },
  {
    path: '/home',
    name: 'Home',
    component: () => import('@/views/Home.vue'),
    meta: { title: '首页', layout: 'default', requiresAuth: false },
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录', layout: 'blank', requiresAuth: false },
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/Register.vue'),
    meta: { title: '注册', layout: 'blank', requiresAuth: false },
  },
  {
    path: '/prompts',
    name: 'Prompts',
    component: () => import('@/views/Prompts.vue'),
    meta: { title: '提示词库', layout: 'default', requiresAuth: true },
  },
  {
    path: '/categories',
    name: 'Categories',
    component: () => import('@/views/Categories.vue'),
    meta: { title: '分类管理', layout: 'default', requiresAuth: true },
  },
  {
    path: '/prompts/new',
    name: 'PromptCreate',
    component: () => import('@/views/PromptCreate.vue'),
    meta: { title: '新建提示词', layout: 'default', requiresAuth: true },
  },
  {
    path: '/prompts/:id(\\d+)',
    name: 'PromptDetail',
    component: () => import('@/views/PromptDetail.vue'),
    meta: { title: '提示词详情', layout: 'default', requiresAuth: true },
  },
  {
    path: '/prompts/:id(\\d+)/tune',
    name: 'PromptTune',
    component: () => import('@/views/PromptTune.vue'),
    meta: { title: 'AI 调优', layout: 'default', requiresAuth: true },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue'),
    meta: { title: '页面不存在', layout: 'default', requiresAuth: false },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 全局路由标题
router.afterEach((to) => {
  const title = (to.meta?.title as string | undefined) || 'PromptHub'
  document.title = title
})

// 鉴权守卫
router.beforeEach(async (to) => {
  // 动态 import store，避免循环依赖
  const { useUserStore } = await import('@/stores/user')
  const userStore = useUserStore()

  const requiresAuth = to.meta?.requiresAuth === true
  const targetName = to.name as string | undefined

  // 鉴权守卫
  const loggedIn = userStore.isLoggedIn()

  // 0) 根路径根据登录态跳转：未登录 → /login；已登录 → /prompts
  if (to.path === '/') {
    return loggedIn
      ? { name: 'Prompts', replace: true }
      : { name: 'Login', replace: true }
  }

  // 1) 受保护路由未登录 → 跳登录
  if (requiresAuth && !loggedIn) {
    const redirect = to.fullPath && /^\/(?!\/)/.test(to.fullPath) ? to.fullPath : undefined
    return {
      name: 'Login',
      query: redirect ? { redirect } : undefined,
      replace: true,
    }
  }

  // 2) 已登录访问 /login 或 /register → 跳提示词库
  if (loggedIn && (targetName === 'Login' || targetName === 'Register')) {
    return { name: 'Prompts', replace: true }
  }

  // 3) 其余放行
  return true
})

export default router