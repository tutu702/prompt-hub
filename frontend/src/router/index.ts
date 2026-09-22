/**
 * Vue Router 路由配置
 * 现阶段仅提供基础骨架，业务页面在后续迭代中按 PRD 逐步添加。
 */
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/home',
  },
  {
    path: '/home',
    name: 'Home',
    component: () => import('@/views/Home.vue'),
    meta: { title: '首页', requiresAuth: false },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue'),
    meta: { title: '页面不存在' },
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

export default router
